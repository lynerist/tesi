package main

import (
	"crypto/md5"
	j "encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type set[T comparable] map[T]bool

func (s set[T])add(o set[T]){
	for k := range o{
		s[k]=true
	}
}

func (s set[T]) String()string {
	if len(s)==0{
		return "empty"
	}
	out := "set{"
	for e := range s{
		out += fmt.Sprintf("%v ", e)
	}
	return out[:len(out)-1]+"}"
}

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

func parseJSON(jsonFile io.Reader)(json map[string]any){
	jsonBin, _ := io.ReadAll(jsonFile)
	j.Unmarshal(jsonBin, &json)

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

func insertVariables(atom any, artifact artifactName, feature featureName, state *State)declaration{
	stringAtom := fmt.Sprint(atom)
	for name, value := range state.attributes[artifact][feature] {
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(value))
	}
	for name := range state.globals.needs[artifact]{
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(state.globals.get(name)))
	}
	return declaration(stringAtom)
}

func getArtifactsFromFeatureJSON(f any)[]any{
	return f.(map[string]any)["artifacts"].([]any)
}

func ifEmptyThenRoot(s featureName)string{
	if s == "" {return "root"}
	return string(s)
}

func generateEdgeData(source, target featureName)map[string]any{
	return map[string]any{"source":ifEmptyThenRoot(source), "target":ifEmptyThenRoot(target)}
}

func generateDependencyEdgeData(source, target featureName, dependencyID int, atom declaration)map[string]any{
	return map[string]any{"source":ifEmptyThenRoot(source), "target":ifEmptyThenRoot(target), 
							"dependencyID":dependencyID, "declaration":atom}
}

func cytoscapeJSON(state *State)([]byte, error){
	var json []any
	for name, feature := range state.features{
		data := map[string]any{"id":ifEmptyThenRoot(name)} //, "parent":ifEmptyThenRoot(*feature.parent)} //REMOVE PARENT??
		classes := []string{"node"}
		if len(feature.artifacts) == 0 {classes = append(classes, "tag")}
		if name == "" {classes = append(classes, "root")}
		json = append(json, map[string]any{"group":"nodes", "data":data, "classes":classes})
		
		/* --- FEATURE MODEL TREE EDGES --- */
		for child := range feature.children{
			json = append(json, map[string]any{"group":"edges", "data":generateEdgeData(name, child)})
		}

		/* --- DEPENDENCIES --- */

		/* ALL */
		var	dependencyID int
		for atom := range feature.requirements.ALL{
			for requiredFeature := range getProviders(atom, state){
				json = append(json, map[string]any{"group":"edges", 
													"data":generateDependencyEdgeData(requiredFeature, name, dependencyID, atom), 
													"classes":[]string{"dependency","dependencyAll"}})
			}
			dependencyID++
		}
		for atom := range feature.variadicRequirements.ALL{
			if feature.requirements.ALL[atom] {continue}
			for requiredFeature := range getProviders(atom, state){
				json = append(json, map[string]any{"group":"edges", 
													"data":generateDependencyEdgeData(requiredFeature, name, dependencyID, atom), 
													"classes":[]string{"dependency","dependencyAll"}})
			}
			dependencyID++
		}

		/* NOT */
		for atom := range feature.requirements.NOT{
			for requiredFeature := range getProviders(atom, state){
				json = append(json, map[string]any{"group":"edges", 
													"data":generateDependencyEdgeData(requiredFeature, name, 0, atom), 
													"classes":[]string{"dependency","dependencyNot"}})
			}
		}
		for atom := range feature.variadicRequirements.NOT{
			if feature.requirements.NOT[atom] {continue}
			for requiredFeature := range getProviders(atom, state){
				json = append(json, map[string]any{"group":"edges", 
													"data":generateDependencyEdgeData(requiredFeature, name, 0, atom), 
													"classes":[]string{"dependency","dependencyNot"}})
			}
		}

		/* ANY */
		done := make(set[string])
		for _, group := range feature.requirements.ANY{
			done[fmt.Sprint(map[declaration]bool(group))] = true
			providers := make(map[featureName]map[string]any)
			for atom := range group{
				for requiredFeature := range getProviders(atom, state){
					if _, ok := providers[requiredFeature]; ok{
						providers[requiredFeature]["data"].(map[string]any)["declaration"] =
						fmt.Sprintf("%s\n%s", providers[requiredFeature]["data"].(map[string]any)["declaration"].(declaration), atom)
					}else{
						providers[requiredFeature] = map[string]any{"group":"edges", 
																	"data":generateDependencyEdgeData(requiredFeature, name, dependencyID, atom), 
																	"classes":[]string{"dependency","dependencyAny"}}
					}

				}
			}
			for _, edge := range providers{
				json = append(json, edge)
			}
			dependencyID++
		}

		for _, group := range feature.variadicRequirements.ANY{
			if done[fmt.Sprint(map[declaration]bool(group))]{continue}
			providers := make(map[featureName]map[string]any)
			for atom := range group{
				for requiredFeature := range getProviders(atom, state){
					if _, ok := providers[requiredFeature]; ok{
						providers[requiredFeature]["data"].(map[string]any)["declaration"] =
						fmt.Sprintf("%s\n%s", providers[requiredFeature]["data"].(map[string]any)["declaration"].(declaration), atom)
					}else{
						providers[requiredFeature] = map[string]any{"group":"edges", 
																	"data":generateDependencyEdgeData(requiredFeature, name, dependencyID, atom), 
																	"classes":[]string{"dependency","dependencyAny"}}
					}

				}
			}
			for _, edge := range providers{
				json = append(json, edge)
			}
			dependencyID++
		}
		//TODO ONE

	}
	return j.MarshalIndent(json, "", "\t")
}

func getProviders(atom declaration, state *State)set[featureName]{
	providers := make(set[featureName])
	providers.add(state.providers[atom])
	providers.add(state.variadicProviders[atom])
	return providers
}

func exportFeatureModelJson(path string, features map[featureName]Feature, state *State){
	outJson, _ := cytoscapeJSON(state)

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(outJson)
}
