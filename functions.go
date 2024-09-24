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

type artifactName	string
type featureName	string
type tagName		string
type variableName	string

func (t tagName) String()string{
	return fmt.Sprintf("T::%s", string(t))
}

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
	if _, ok := artifact["globals"]; !ok {artifact["globals"]=[]any{}}
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

func insertVariables(atom any, artifact artifactName, feature featureName, variables map[artifactName]map[featureName]map[variableName]any, globals globalRegister)string{
	stringAtom := fmt.Sprint(atom)
	for name, value := range variables[artifact][feature] {
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(value))
	}
	for name := range globals.needs[artifact]{
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(globals.get(name)))
	}
	return stringAtom
}

type Artifact map[string]any 

func (a Artifact) name() artifactName{
	return artifactName(a["name"].(string))
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

func getArtifactsFromFeatureJSON(f any)[]any{
	return f.(map[string]any)["artifacts"].([]any)
}
func getNameFromFeatureJSON(f any) featureName{
	return featureName(f.(map[string]any)["name"].(string))
}

func stringToAN(s any)artifactName{
	return artifactName(s.(string))
}

type globalRegister struct {
	proposed map[variableName]map[any]int
	elected map[variableName]any
	needs map[artifactName]map[variableName]bool
}
func newGlobalRegister()(newRegister globalRegister){
	newRegister.proposed = make(map[variableName]map[any]int)
	newRegister.elected = make(map[variableName]any)
	newRegister.needs = make(map[artifactName]map[variableName]bool)
	return
}
func (gr globalRegister)put(name variableName, value any, artifact artifactName){
	if len(gr.proposed[name])==0{
		gr.proposed[name] = make(map[any]int)
	} 
	gr.proposed[name][value]++
	if gr.proposed[name][value] > gr.proposed[name][gr.elected[name]] || len(gr.proposed)==1{
		gr.elected[name]=value
	}
	if len(gr.needs[artifact])==0{
		gr.needs[artifact] = make(map[variableName]bool)
	}
	gr.needs[artifact][name]=true
}
func (gr globalRegister)get(name variableName)any{
	return gr.elected[name]
}

type Feature struct {
	name featureName
	artifacts []artifactName
	tags map[tagName]bool
	children map[featureName]bool
}

func newFeatureName(feature any, id int)featureName{
	return featureName(fmt.Sprintf("%s::%d", feature.(map[string]any)["name"].(string), id))
}

func newFeature(name featureName, composingArtifacts []any, artifacts map[artifactName]Artifact)Feature{
	feature := Feature{name, nil, make(map[tagName]bool), make(map[featureName]bool)}

	for _, artifact := range composingArtifacts {
		feature.artifacts = append(feature.artifacts, artifactName(artifact.(string)))
		for _, tag := range artifacts[stringToAN(artifact)].tags(){
			feature.tags[tagName(tag.(string))] = true
		}
	}
	return feature
}

func newAbstractFeature(name featureName)Feature{
	return Feature{name, nil, nil, make(map[featureName]bool)}
}

func (f Feature) String()string {
	var tags []tagName
	for tag := range f.tags{
		tags = append(tags, tag)
	}
	var children []featureName
	for child := range f.children{
		children = append(children, child)
	}
	return fmt.Sprintf("'%s' --> artifacts:%v tags: %v children: %v", f.name, f.artifacts, tags, children)
}

func (f Feature) isAbstract()bool{
	return len(f.artifacts)>0
}

func generateFeatureTree(root featureName, features map[featureName]Feature){
	tagCount := make(map[tagName]int)
	for child := range features[root].children{
		for tag := range features[child].tags{
			tagCount[tag]++
		}
	}

	for{
		var mostPresentTag tagName 
		for tag, count := range tagCount{
			if count > tagCount[mostPresentTag] && count>1{
				mostPresentTag = tag
			}
		}
		if mostPresentTag == ""{
			break
		}

		newAbstractFeature := newAbstractFeature(featureName(fmt.Sprintf("%s::%d",mostPresentTag, len(features))))
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
		features[root].children[newAbstractFeature.name] = true
		features[newAbstractFeature.name] = newAbstractFeature
	}
	for child := range features[root].children{
		generateFeatureTree(child, features)
	}
}

func printTree(root featureName, indent int, features map[featureName]Feature){
	if len(features[root].children)==0{
		fmt.Printf("%s%s\n", strings.Repeat("\t", indent), root)
		return
	}
	fmt.Printf("%s%s -> [\n", strings.Repeat("\t", indent), root)
	for child := range features[root].children{
		printTree(child, indent+1, features)
	}
	fmt.Printf("%s]\n", strings.Repeat("\t", indent))
}