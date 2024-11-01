package main

import (
	"crypto/md5"
	j "encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type set[T comparable] map[T]bool //TODO MAKE IT to struct{}???
type valueOrSet[T comparable] interface{}

func (s set[T]) add (toAdd valueOrSet[T]){
	if s == nil {
		panic("insert in nil set")
	}
	switch toAdd := toAdd.(type) {
	case T:
		s[toAdd]=true
	case set[T]:
		for k := range toAdd{
			s[k]=true
		}
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
	for name, value := range state.variables[artifact][feature] {
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(value))
	}
	for name := range state.globals.neededByArtifact[artifact]{
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

func getVariables(feature *Feature, state *State)map[string]variableValue{
	attributes := make(map[string]variableValue)
	for artifact := range feature.artifacts{
		if state.artifacts[artifact].isVariadic(){
			for variable, value := range state.variables[artifact][feature.name]{
				attributes[fmt.Sprintf("%s%s", artifact,variable)] = value
			}
		}
	}
	return attributes	
}

func getGlobals(feature *Feature, state *State)map[variableName]variableValue{
	globals := make(map[variableName]variableValue)
	for artifact := range feature.artifacts{
		if state.artifacts[artifact].isVariadic(){
			for global := range state.globals.neededByArtifact[artifact]{
				globals[global] = state.globals.get(global)
			}
		}
	}
	return globals	
}

func countLevels(feature featureName, level int, levels map[int]set[featureName], state *State){
	if levels[level] == nil {levels[level] = make(set[featureName])}
	levels[level].add(feature)
	for child := range state.features[feature].children{
		countLevels(child, level+1, levels, state)
	}
}

func addClassInExtractedJSON(json []map[string]any, index int, class string){
	json[index]["classes"]=append(json[index]["classes"].([]string),class)
}

func extractCytoscapeJSON(state *State)([]byte, error){
	featuresIndexes := make(map[featureName]int) //to get the index of a specific feature in the json
	var json []map[string]any 
	for name, feature := range state.features{
		featuresIndexes[feature.name] = len(json)
		/* --- --> NODE <-- --- */
		// node id && attributes
		data := map[string]any{"id":ifEmptyThenRoot(name), 
								"variables":getVariables(&feature, state), 
								"globals":getGlobals(&feature,state)}

		classes := []string{"node"}
		if len(feature.artifacts) == 0 {
			classes = append(classes, "tag")
			data["abstract"]=true
			if name == ROOT {classes = append(classes, "root")}
		}

		// append new node
		json = append(json, map[string]any{"group":"nodes", "data":data, "classes":classes})
		
		/* --- --> EDGES <-- --- */

		/* --- FEATURE MODEL TREE --- */
		for child := range feature.children{
			json = append(json, map[string]any{"group":"edges", "data":generateEdgeData(name, child)})
		}

		/* --- DEPENDENCIES --- */
		requirements := feature.getRequirements(state)

		/* ALL */
		var	dependencyID int

		for atom := range requirements.ALL{
			if providers := getProviders(atom, state, feature.name); len(providers)>0{
				for requiredFeature := range providers{
					json = append(json, map[string]any{"group":"edges", 
														"data":generateDependencyEdgeData(requiredFeature, name, dependencyID, atom), 
														"classes":[]string{"dependency","dependencyAll"}})
				}
			}else{
				addClassInExtractedJSON(json, featuresIndexes[feature.name], "deadFeature")
			}
			dependencyID++
		}
		
		/* NOT */
		for atom := range requirements.NOT{
			for requiredFeature := range getProviders(atom, state, feature.name){
				json = append(json, map[string]any{"group":"edges", 
													"data":generateDependencyEdgeData(requiredFeature, name, 0, atom), 
													"classes":[]string{"dependency","dependencyNot"}})
			}
		}

		/* ANY */
		for _, group := range *requirements.ANY{
			providers := make(map[featureName]map[string]any)
			for atom := range group{
				for requiredFeature := range getProviders(atom, state, feature.name){
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
			if len(providers)==0{
				addClassInExtractedJSON(json, featuresIndexes[feature.name], "deadFeature")
			}
			dependencyID++
		}

		/* ONE */

		for _, group := range *requirements.ONE{
			providers := make(map[featureName]map[string]any)
			for atom := range group{
				for requiredFeature := range getProviders(atom, state, feature.name){
					if _, ok := providers[requiredFeature]; ok{
						providers[requiredFeature]["data"].(map[string]any)["declaration"] =
						fmt.Sprintf("%s\n%s", providers[requiredFeature]["data"].(map[string]any)["declaration"].(declaration), atom)
					}else{
						providers[requiredFeature] = map[string]any{"group":"edges", 
																	"data":generateDependencyEdgeData(requiredFeature, name, dependencyID, atom), 
																	"classes":[]string{"dependency","dependencyOne"}}
					}
				}
			}
			for _, edge := range providers{
				json = append(json, edge)
			}
			if len(providers)==0{
				addClassInExtractedJSON(json, featuresIndexes[feature.name], "deadFeature")
			}
			dependencyID++
		}
	}

	// MOVE GLOBALS IN MOST UPPER COMMON ANCESTOR

	levels := make(map[int]set[featureName])
	countLevels(ROOT, 0, levels, state)

	for level := len(levels)-1; level >=0; level--{
		// Count how many times each global appears. If one appear more then one time it has to be moved to the upper node.
		globalsCount := make(map[variableName]int)
		for feature := range state.features{
			for global := range json[featuresIndexes[feature]]["data"].(map[string]any)["globals"].(map[variableName]variableValue) {
				globalsCount[global]++
			}
		}
		for feature := range levels[level]{
			toMove := make(set[variableName])
			for global := range json[featuresIndexes[feature]]["data"].(map[string]any)["globals"].(map[variableName]variableValue) {
				if globalsCount[global]>1{
					toMove.add(global)
				}
			}
			for global := range toMove{
				json[featuresIndexes[*state.features[feature].parent]]["data"].(map[string]any)["globals"].(map[variableName]variableValue)[global] = state.globals.get(global)
				delete(json[featuresIndexes[feature]]["data"].(map[string]any)["globals"].(map[variableName]variableValue), global)
			}
		}
	}

	return j.MarshalIndent(json, "", "\t")
}

func getProviders(atom declaration, state *State, requier featureName)set[featureName]{
	providers := make(set[featureName])
	
	for possibleProvider := range state.possibleProviders[atom]{
		if state.features[possibleProvider].isProviding(atom, state){
			providers.add(possibleProvider)
		}
	}

	delete(providers, requier)
	return providers
}


func updatePossibleProvidersByVariableChange(artifact artifactName, feature featureName, state *State){
	for provided := range state.features[feature].provisions[artifact]{
		atom := insertVariables(provided, artifact, feature, state)
		if _, ok := state.possibleProviders[atom]; !ok {
			state.possibleProviders[atom] = make(set[featureName])
		}
		state.possibleProviders[atom].add(feature)
	}
}

func updatePossibleProvidersByGlobalChange(global variableName, state *State){
	for feature := range state.globals.usedBy[global]{
		for artifact := range state.features[feature].artifacts{
			for provided := range state.features[feature].provisions[artifact]{
				atom := insertVariables(provided, artifact, feature, state)
				if _, ok := state.possibleProviders[atom]; !ok {
					state.possibleProviders[atom] = make(set[featureName])
				}
				state.possibleProviders[atom].add(feature)
			}
		}
	}
}

func exportFeatureModelJson(path string, state *State){
	outJson, _ := extractCytoscapeJSON(state)

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(outJson)
}
