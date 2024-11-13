package main

import (
	"crypto/md5"
	j "encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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

func parseJSON(jsonFile io.Reader)(json map[string]any){
	jsonBin, _ := io.ReadAll(jsonFile)
	j.Unmarshal(jsonBin, &json)

	for k := range json["artifacts"].([]any){
		fillMissingKeys((json["artifacts"].([]any)[k]).(map[string]any))
	}
	return
}

//fill missing keys in the json with empty value to prevent wrong type assertions
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
	if _, ok := artifact["variables"]; !ok {artifact["variables"]=[]any{}}
	if _, ok := artifact["globals"]; !ok {artifact["globals"]=[]any{}}
	if _, ok := artifact["tags"]; !ok {artifact["tags"]=[]string{}}

	if _, ok := artifact["conditionalProvides"]; !ok {artifact["conditionalProvides"]=[]any{}}
}

func md5hash(s string)hash{
	return hash(fmt.Sprintf("'%x'", md5.Sum([]byte(s))))
}

func unHash(hashedVar hash, state *State)any{
	var answer any
	answer, ok := state.hashesToDeclaration[hashedVar]
	if !ok {
		answer, ok = state.hashesToFeature[hashedVar]
		if !ok{
			answer = fmt.Sprint(hashedVar)
		}
	}
	return answer
}

func insertAttributes(atom any, artifact artifactName, feature featureName, state *State)declaration{
	stringAtom := fmt.Sprint(atom)
	for name, value := range state.variables[artifact][feature] {
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(value))
	}
	for name := range state.artifacts[artifact].globalsDefault{
		stringAtom = strings.ReplaceAll(stringAtom, string(name), fmt.Sprint(state.globals.get(name)))
	}
	return declaration(stringAtom)
}

func getArtifactsFromFeatureJSON(f any)[]any{
	return f.(map[string]any)["artifacts"].([]any)
}

func generateEdgeData(source, target featureName)map[string]any{
	return map[string]any{"source":source, "target":target}
}

func generateDependencyEdgeData(source, target featureName, dependencyID int, atom declaration)map[string]any{
	return map[string]any{"source":source, "target":target, 
							"dependencyID":dependencyID, "declaration":atom}
}

//get all the couples variable-value used by a feature
func getVariables(feature *Feature, state *State)map[string]attributeValue{
	attributes := make(map[string]attributeValue)
	for artifact := range feature.artifacts{
		for variable, value := range state.variables[artifact][feature.name]{
			attributes[fmt.Sprintf("%s%s", artifact,variable)] = value
		}
	}
	return attributes	
}

//get all the couples global-value used by a feature
func getGlobals(feature *Feature, state *State)map[attributeName]attributeValue{
	globals := make(map[attributeName]attributeValue)
	for artifact := range feature.artifacts{
		for global := range state.artifacts[artifact].globalsDefault{
			globals[global] = state.globals.get(global)
		}
	}
	return globals	
}

//recursive function to count the deepness of a feature in the feature model tree
func countLevels(feature featureName, level int, levels map[int]set[featureName], state *State){
	if levels[level] == nil {levels[level] = make(set[featureName])}
	levels[level].add(feature)
	for child := range state.features[feature].children{
		countLevels(child, level+1, levels, state)
	}
}

func handleDeadFeature(json []map[string]any, index int, state *State, feature featureName){
	state.deadFeatures.add(feature)
	state.activeFeatures.remove(feature)  
	json[index]["data"].(map[string]any)["deadFeature"] = true
	json[index]["data"].(map[string]any)["active"] = false
}

func checkDeadFeaturesANDextractInterfaceJSON(state *State)([]byte, error){
	state.deadFeatures = make(set[featureName])

	featuresIndexes := make(map[featureName]int) //to get the index of a specific feature in the json
	var json []map[string]any 

	for name, feature := range state.features{
		featuresIndexes[feature.name] = len(json)
		/* --- --> NODE <-- --- */
		// node id && attributes
		data := map[string]any{"id":name, 
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
				handleDeadFeature(json, featuresIndexes[feature.name], state, feature.name)
			}
			dependencyID++
		}
		
		/* NOT */
		for atom := range requirements.NOT{
			for requiredNotFeature := range getProviders(atom, state, feature.name){
				json = append(json, map[string]any{"group":"edges", 
													"data":generateDependencyEdgeData(requiredNotFeature, name, 0, atom), 
													"classes":[]string{"dependency","dependencyNot"}})
				if state.isActive(requiredNotFeature) {
					handleDeadFeature(json, featuresIndexes[feature.name], state, feature.name)
				}
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
				handleDeadFeature(json, featuresIndexes[feature.name], state, feature.name)
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
				handleDeadFeature(json, featuresIndexes[feature.name], state, feature.name)
			}
			dependencyID++
		}
	}

	// MOVE GLOBALS IN MOST UPPER COMMON ANCESTOR
	levels := make(map[int]set[featureName])
	countLevels(ROOT, 0, levels, state)

	for level := len(levels)-1; level >=0; level--{
		// Count how many times each global appears. If one appear more then one time it has to be moved to the upper node.
		globalsCount := make(map[attributeName]int)
		for feature := range state.features{
			for global := range json[featuresIndexes[feature]]["data"].(map[string]any)["globals"].(map[attributeName]attributeValue) {
				globalsCount[global]++
			}
		}
		for feature := range levels[level]{
			toMove := make(set[attributeName])
			for global := range json[featuresIndexes[feature]]["data"].(map[string]any)["globals"].(map[attributeName]attributeValue) {
				if globalsCount[global]>1{
					toMove.add(global)
				}
			}
			for global := range toMove{
				json[featuresIndexes[*state.features[feature].parent]]["data"].(map[string]any)["globals"].(map[attributeName]attributeValue)[global] = state.globals.get(global)
				delete(json[featuresIndexes[feature]]["data"].(map[string]any)["globals"].(map[attributeName]attributeValue), global)
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

//Whenever a variable change value, declarations provided by the feature that's using that variable may change so we have to update them
func updatePossibleProvidersByVariableChange(artifact artifactName, feature featureName, state *State){
	for provided := range state.features[feature].provisions[artifact]{
		atom := insertAttributes(provided, artifact, feature, state)
		if _, ok := state.possibleProviders[atom]; !ok {
			state.possibleProviders[atom] = make(set[featureName])
		}
		state.possibleProviders[atom].add(feature)
	}
}

//Whenever a global change value, declarations provided by the features that are using that global may change so we have to update them
func updatePossibleProvidersByGlobalChange(global attributeName, state *State){
	for feature := range state.globals.usedBy[global]{
		for artifact := range state.features[feature].artifacts{
			for provided := range state.features[feature].provisions[artifact]{
				atom := insertAttributes(provided, artifact, feature, state)
				if _, ok := state.possibleProviders[atom]; !ok {
					state.possibleProviders[atom] = make(set[featureName])
				}
				state.possibleProviders[atom].add(feature)
			}
		}
	}
}

func exportFeatureModelJson(path string, state *State){
	outJson, _ := checkDeadFeaturesANDextractInterfaceJSON(state)

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(outJson)
}

func findProvidersForSelectedDeclarations(invalidFeatureRequirements map[featureName]Requirements, state *State)map[declaration]set[featureName]{
	providers := make(map[declaration]set[featureName])
	for feature, requirements := range invalidFeatureRequirements{
		for atom := range requirements.ALL{
			if _, exists := providers[atom]; !exists {providers[atom]=make(set[featureName])}
			providers[atom].add(getProviders(atom, state, feature))
		}
		for atom := range requirements.NOT{
			if _, exists := providers[atom]; !exists {providers[atom]=make(set[featureName])}
			providers[atom].add(getProviders(atom, state, feature))
		}
		for _, group := range *requirements.ANY{
			for atom := range group {
				if _, exists := providers[atom]; !exists {providers[atom]=make(set[featureName])}
				providers[atom].add(getProviders(atom, state, feature))
			}
		}
		for _, group := range *requirements.ONE{
			for atom := range group {
				if _, exists := providers[atom]; !exists {providers[atom]=make(set[featureName])}
				providers[atom].add(getProviders(atom, state, feature))
			}
		}
	}
	return providers
}