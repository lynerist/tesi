package main

import (
	"crypto/md5"
	j "encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"net/http"
)

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

func provides(who, what any, hashes map[string]string)string{
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

func insertVariables(atom any, artifact artifactName, feature featureName, variables map[artifactName]map[featureName]map[variableName]any, globals globalContext)string{
	stringAtom := fmt.Sprint(atom)
	for name, value := range variables[artifact][feature] {
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(value))
	}
	for name := range globals.needs[artifact]{
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(globals.get(name)))
	}
	return stringAtom
}

func getArtifactsFromFeatureJSON(f any)[]any{
	return f.(map[string]any)["artifacts"].([]any)
}

func ifEmptyThenRoot(s featureName)string{
	if s == "" {return "root"}
	return string(s)
}

func cytoscapeJSON(features map[featureName]Feature)([]byte, error){
	var json []any
	for name, feature := range features{
		data := map[string]any{"id":ifEmptyThenRoot(name)} //, "parent":ifEmptyThenRoot(*feature.parent)} //REMOVE PARENT??
		classes := []string{"node"}
		if len(feature.artifacts) == 0 {classes = append(classes, "tag")}
		if name == "" {classes = append(classes, "root")}
		json = append(json, map[string]any{"group":"nodes", "data":data, "classes":classes})
		for child := range feature.children{
			data := map[string]any{"source":ifEmptyThenRoot(name), "target":ifEmptyThenRoot(child)}
			json = append(json, map[string]any{"group":"edges", "data":data})
		}
	}
	return j.MarshalIndent(json, "", "\t")
}

func exportFeatureModelJson(path string, features map[featureName]Feature){
	outJson, _ := cytoscapeJSON(features)

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(outJson)
}

func initLocalServer(){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "interface/" + r.URL.Path[1:])
	})
	
	fmt.Println(http.ListenAndServe(":"+PORT, nil))
}