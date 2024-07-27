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

const VERBOSE = false

func readJSON(fileName string)(json []map[string]any){
	jsonFile, _ := os.Open(fileName)
	jsonBin, _ := io.ReadAll(jsonFile)
	j.Unmarshal(jsonBin, &json)
	jsonFile.Close()
	return
}

func fillMissingKeys(artifact map[string]any){
	if _, ok := artifact["requires"]; !ok {artifact["requires"]=make(map[string]any)}
	if _, ok := artifact["requires"].(map[string]any)["all"]; !ok {
		artifact["requires"].(map[string]any)["all"] = []any{}
	}
	if _, ok := artifact["requires"].(map[string]any)["not"]; !ok {
		artifact["requires"].(map[string]any)["not"] = []any{}
	}
	if _, ok := artifact["requires"].(map[string]any)["any"]; !ok {
		artifact["requires"].(map[string]any)["any"] = [][]any{}
	}
	if _, ok := artifact["requires"].(map[string]any)["one"]; !ok {
		artifact["requires"].(map[string]any)["one"] = [][]any{}
	}
	if _, ok := artifact["provides"]; !ok {artifact["provides"]=[]any{}}
	if _, ok := artifact["conditionalProvides"]; !ok {artifact["conditionalProvides"]=[]any{}}
}

func md5hash(s string)string{
	return fmt.Sprintf("'%x'", md5.Sum([]byte(s)))
}

func requireAll(who, what any, hashes map[string]string)string{
	requiredHash := md5hash(fmt.Sprint(what))
	hashes[requiredHash] = fmt.Sprint(what)
	requiringHash := md5hash(fmt.Sprint(who))
	hashes[requiringHash] = fmt.Sprint(who)
	return fmt.Sprintf("requiresAll(%s,%s).", requiringHash,requiredHash)
}

func requireNot(who, what any, hashes map[string]string)string{
	requiredHash := md5hash(fmt.Sprint(what))
	hashes[requiredHash] = fmt.Sprint(what)
	requiringHash := md5hash(fmt.Sprint(who))
	hashes[requiringHash] = fmt.Sprint(who)
	return fmt.Sprintf("requiresNot(%s,%s).", requiringHash,requiredHash)
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
	return fmt.Sprintf("%s\n%s\n%s\n%s", 
						core.program["start"],
						core.program["requiresAll"],
						core.program["requiresNot"],
						core.program["provides"])
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
	return prologCore{prolog.New(nil,nil), map[string]string{"start":program}}
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

				toPrint = append(toPrint, fmt.Sprintf("%s:%s\t",k,hashes["'"+fmt.Sprint(v)+"'"]))
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