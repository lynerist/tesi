package main

import (
	"bufio"
	"crypto/md5"
	j "encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"github.com/ichiban/prolog"
)

const VERBOSE = true
const ALL 		string = "all"
const NOT 		string = "not"
const ANY 		string = "any"
const ONE 		string = "one"
const REQUIRES 	string = "requires"
const PROVIDES 	string = "provides"


func readJSON(fileName string)(json map[string]any){
	jsonFile, _ := os.Open(fileName)
	jsonBin, _ := io.ReadAll(jsonFile)
	j.Unmarshal(jsonBin, &json)
	jsonFile.Close()

	for k := range json["artifacts"].([]any){
		fillMissingKeys((json["artifacts"].([]any)[k]).(map[string]any))
	}

	return
}

func fillMissingKeys(artifact map[string]any){
	if _, ok := artifact[REQUIRES]; !ok {artifact[REQUIRES]=make(map[string]any)}
	if _, ok := artifact[REQUIRES].(map[string]any)[ALL]; !ok {
		artifact[REQUIRES].(map[string]any)[ALL] = []any{}
	}
	if _, ok := artifact[REQUIRES].(map[string]any)[NOT]; !ok {
		artifact[REQUIRES].(map[string]any)[NOT] = []any{}
	}
	if _, ok := artifact[REQUIRES].(map[string]any)[ANY]; !ok {
		artifact[REQUIRES].(map[string]any)[ANY] = make([]any, 0)
	}
	if _, ok := artifact[REQUIRES].(map[string]any)[ONE]; !ok {
		artifact[REQUIRES].(map[string]any)[ONE] = make([]any, 0)
	}

	if _, ok := artifact[PROVIDES]; !ok {artifact[PROVIDES]=[]any{}}
	if _, ok := artifact["attributes"]; !ok {artifact["attributes"]=[]any{}}
	if _, ok := artifact["tags"]; !ok {artifact["tags"]=[]string{}}

	if _, ok := artifact["conditionalProvides"]; !ok {artifact["conditionalProvides"]=[]any{}}
}

func md5hash(s string)string{
	return fmt.Sprintf("'%x'", md5.Sum([]byte(s)))
}

func calculateAndAddHashes(who, what any, hashes map[string]string)(string, string){
	requiredHash := md5hash(fmt.Sprint(what)); hashes[requiredHash] = fmt.Sprint(what)
	requiringHash := md5hash(fmt.Sprint(who)); hashes[requiringHash] = fmt.Sprint(who)
	return requiringHash, requiredHash
}

func requiresAll(who, what any, hashes map[string]string)string{
	requiringHash,requiredHash := calculateAndAddHashes(who,what,hashes)
	return fmt.Sprintf("requiresAll(%s,%s).", requiringHash,requiredHash)
}

func requiresNot(who, what any, hashes map[string]string)string{
	requiringHash,requiredHash := calculateAndAddHashes(who,what,hashes)
	return fmt.Sprintf("requiresNot(%s,%s).", requiringHash,requiredHash)
}

func requiresAny(who, what any, groupID int, hashes map[string]string)string{
	requiringHash,requiredHash := calculateAndAddHashes(who,what,hashes)
	return fmt.Sprintf("requiresAny(%s,%s,%d).", requiringHash,requiredHash, groupID)
}

func requiresOne(who, what any, groupID int, hashes map[string]string)string{
	requiringHash,requiredHash := calculateAndAddHashes(who,what,hashes)
	return fmt.Sprintf("requiresOne(%s,%s,%d).", requiringHash,requiredHash, groupID)
}

func provide(who, what any, hashes map[string]string)string{
	providedHash := md5hash(fmt.Sprint(what))
	hashes[providedHash] = fmt.Sprint(what)
	providingHash := md5hash(fmt.Sprint(who))
	hashes[providingHash] = fmt.Sprint(who)
	return fmt.Sprintf("provides(%s,%s).", providingHash,providedHash)
}

func log(s ...any){
	if (VERBOSE){
		fmt.Println(s...)
	}
}

type prologCore struct{
	interpreter *prolog.Interpreter
	program map[string]string
}

func (core *prologCore) addLine(s, where string){
	core.program[where] += "\n"+s
}

func (core *prologCore)getProgram()string{
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", 
						core.program["start"],
						core.program["requiresAll"],
						core.program["requiresNot"],
						core.program["requiresAny"],
						core.program["requiresOne"],
						core.program[PROVIDES])
}

func (core prologCore) runProgram(){
	err:=core.interpreter.Exec(core.getProgram())
	if err != nil{
		fmt.Println("Error in the program:", err)
	}
}

func setupProlog() prologCore{
	file, _ := os.Open("core.pl")
	sc := bufio.NewScanner(file)
	var program string
	for sc.Scan(){
		program += sc.Text() + "\n"
	}
	return prologCore{prolog.New(nil,nil), map[string]string{
											"start":program, 
											"requiresAll":"requiresAll(foo,foo).",
											"requiresNot":"requiresNot(foo,foo).",
											"requiresAny":"requiresAny(foo,foo,0).",
											"requiresOne":"requiresOne(foo,foo,0).",
											}}
}

var prologErrorsMeaning = map[string]string {
	"EOF":"Missing end of the query.",
}

func prologQueryConsole(core prologCore, hashes map[string]string){
	sc := bufio.NewScanner(os.Stdin)
	for fmt.Print("?- "); sc.Scan(); fmt.Print("?- "){
		query := sc.Text()
		for hash, fullName := range hashes{
			query = strings.ReplaceAll(query, fullName, hash)
		}

		solutions, err := core.interpreter.Query(query)
		if err != nil{
			if meaning, ok := prologErrorsMeaning[fmt.Sprint(err)]; ok{
				fmt.Printf("Errore in '%s': %s\n\n", query, meaning)
			}else{
				fmt.Printf("Errore in '%s': %v\n\n", query, err)
			}
			continue
		}

		var output string
		for solutions.Next(){
			variables := make(map[string]any)
			solutions.Scan(variables)
			if len(variables) == 0{
				output += "true."
			}
			toPrint := make([]string,0,len(output))
			for k,v := range variables{
				answer, ok := hashes["'"+fmt.Sprint(v)+"'"]
				if !ok {answer = fmt.Sprint(v)}
				toPrint = append(toPrint, fmt.Sprintf("%s:%s\t",k,answer))
			}
			sort.Strings(toPrint)
			output += strings.Join(toPrint, "\t") + "\n"
		}
		
		if output == "" {
			fmt.Print("false.\n")
		}
		fmt.Println(output)
	}
}

func insertVariables(atom any, artifact, feature string, variables map[string]map[string]variableValue, globals globalRegister)string{
	stringAtom := fmt.Sprint(atom)
	for name, value := range variables[artifact][feature] {
		stringAtom = strings.ReplaceAll(stringAtom, name, fmt.Sprint(value))
	}
	for name := range globals.needs[artifact]{
		stringAtom = strings.ReplaceAll(stringAtom, name, fmt.Sprint(globals.get(name)))
	}
	return stringAtom
}

type Artifact map[string]any 

func (a Artifact) name() string{
	return a["name"].(string)
}
func (a Artifact) requires(how string)[]any{
	return a[REQUIRES].(map[string]any)[how].([]any)
} 
func (a Artifact) provides()[]any{
	return a[PROVIDES].([]any)
} 
func (a Artifact) attributes()[]any{
	return a["attributes"].([]any)
} 
func (a Artifact) globals()[]any{
	return a["globals"].([]any)
} 
func (a Artifact) tags()[]any{
	return a["tags"].([]any)
}

type FeatureArtifacts map[string]any

func (f FeatureArtifacts) name() string{
	return f["name"].(string)
}
func (f FeatureArtifacts) artifacts()[]any{
	return f["artifacts"].([]any)
}

type variableValue map[string]any

type globalRegister struct {
	proposed map[string]map[any]int
	elected map[string]any
	needs map[string]map[string]bool
}
func newGlobalRegister()(newRegister globalRegister){
	newRegister.proposed = make(map[string]map[any]int)
	newRegister.elected = make(map[string]any)
	newRegister.needs = make(map[string]map[string]bool)
	return
}
func (gr globalRegister)put(name string, value any, artifact string){
	if len(gr.proposed[name])==0{
		gr.proposed[name] = make(map[any]int)
	} 
	gr.proposed[name][value]++
	if gr.proposed[name][value] > gr.proposed[name][gr.elected[name]] || len(gr.proposed)==1{
		gr.elected[name]=value
	}
	if len(gr.needs[artifact])==0{
		gr.needs[artifact] = make(map[string]bool)
	}
	gr.needs[artifact][name]=true
}
func (gr globalRegister)get(name string)any{
	return gr.elected[name]
}

type Feature struct {
	name string
	artifacts []string
	tags map[string]bool
	children map[string]bool
}

func newFeature(name string, fa FeatureArtifacts, artifacts map[string]Artifact)Feature{
	feature := Feature{name, nil, make(map[string]bool), make(map[string]bool)}

	for _, artifact := range fa.artifacts(){
		feature.artifacts = append(feature.artifacts, artifact.(string))
		for _, tag := range artifacts[artifact.(string)].tags(){
			feature.tags[tag.(string)] = true
		}
	}
	return feature
}

func newAbstractFeature(name string)Feature{
	return Feature{name, nil, nil, make(map[string]bool)}
}

func (f Feature) String()string {
	var tags []string
	for tag := range f.tags{
		tags = append(tags, tag)
	}
	var children []string
	for child := range f.children{
		children = append(children, child)
	}
	return fmt.Sprintf("'%s' --> artifacts:%v tags: %v children: %v", f.name, f.artifacts, tags, children)
}

func (f Feature) isAbstract()bool{
	return len(f.artifacts)>0
}

func generateFeatureTree(root string, features map[string]Feature){
	tagCount := make(map[string]int)
	for child := range features[root].children{
		for tag := range features[child].tags{
			tagCount[tag]++
		}
	}

	for{
		var mostPresentTag string 
		for tag, count := range tagCount{
			if count > tagCount[mostPresentTag] && count>1{
				mostPresentTag = tag
			}
		}
		if mostPresentTag == ""{
			break
		}

		newAbstractFeature := newAbstractFeature(fmt.Sprintf("%s::%d",mostPresentTag, len(features)))
		for child := range features[root].children{
			if features[child].tags[mostPresentTag]{
				newAbstractFeature.children[child] = true
			}
		}
		for child := range newAbstractFeature.children{
			for tag := range features[child].tags{
				tagCount[tag]--
			}
			delete(features[child].tags, mostPresentTag)
			delete(features[root].children, child)
		}
		features[root].children[mostPresentTag] = true
		features[mostPresentTag] = newAbstractFeature
	}
	for child := range features[root].children{
		generateFeatureTree(child, features)
	}
}