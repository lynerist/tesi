package main

import (
	"fmt"
)

func main(){

	/* --- READ JSON ---*/
	jsonName := "domotica"
	json := readJSON(fmt.Sprintf("json/%s.JSON", jsonName))
	hashesToText := make(map[string]string)
	core := setupProlog()
	
	/* --- MAP ARTIFACTS TO FEATURES AND FEATURES TO ARTIFACTS ---*/

	artifactsInFeature := make(map[featureName][]artifactName)
	featureWithArtifact := make(map[artifactName][]featureName)
	for _, f := range json["features"].([]any){
		for _, artifact := range getArtifactsFromFeatureJSON(f){
			artifactsInFeature[getNameFromFeatureJSON(f)]=append(artifactsInFeature[getNameFromFeatureJSON(f)], stringToAN(artifact))
			featureWithArtifact[stringToAN(artifact)]=append(featureWithArtifact[stringToAN(artifact)], getNameFromFeatureJSON(f))
		}
	}
	
	/* --- STORE ARTIFACTS ---*/

	artifacts := make(map[artifactName]Artifact)
	attributes := make(map[artifactName]map[featureName]variableValue)
	globals := newGlobalRegister()
	
	for _, a := range json["artifacts"].([]any){
		artifact := Artifact(a.(map[string]any))
		artifacts[artifact.name()] = artifact

		/* --- STORE ATTRIBUTES ---*/ 		
		for _, attribute := range artifact.attributes(){
			if name, ok := attribute.(map[string]any)["name"]; ok && []rune(name.(string))[0]=='$'{
				if value, ok := attribute.(map[string]any)["default"]; ok{
					attributes[artifact.name()] = make(map[featureName]variableValue)
					for _, feature := range featureWithArtifact[artifact.name()]{
						attributes[artifact.name()][feature] = make(variableValue) 
						attributes[artifact.name()][feature][name.(string)] = value
					}
					attributes[artifact.name()][""] = make(variableValue) 
					attributes[artifact.name()][""][name.(string)] = value
				}
			}
		}

		/* --- STORE GLOBALS ---*/ 		
		for _, global := range artifact.globals(){
			if name, ok := global.(map[string]any)["name"]; ok && []rune(name.(string))[0]=='@'{
				if value, ok := global.(map[string]any)["default"]; ok{
					globals.put(name.(string), value, artifact.name())
				}
			}
		}

		log(artifact.name())
		
		for i, required := range artifact.requires(ALL){
			if i==0 {log("requires all:")}
			required = insertVariables(required, artifact.name(), "", attributes, globals)
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresAll(artifact.name(),required,hashesToText), "requiresAll")
		}
		for i, required := range artifact.requires(NOT){
			if i==0 {log("requires not:")}
			required = insertVariables(required, artifact.name(), "", attributes, globals)
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresNot(artifact.name(),required,hashesToText), "requiresNot")
		}

		for groupID, requiredAnyGroup := range artifact.requires(ANY){
			log("\t","any of:",requiredAnyGroup)
			for _, required :=  range requiredAnyGroup.([]any){
				required = insertVariables(required, artifact.name(), "", attributes, globals)
				log("\t\t",required,md5hash(fmt.Sprint(required)))
				core.addLine(requiresAny(artifact.name(), required, groupID, hashesToText), "requiresAny")
			}
		}

		for groupID, requiredAnyGroup := range artifact.requires(ONE){
			log("\t","one of:",requiredAnyGroup)
			for _, required :=  range requiredAnyGroup.([]any){
				required = insertVariables(required, artifact.name(), "", attributes, globals)
				log("\t\t",required,md5hash(fmt.Sprint(required)))
				core.addLine(requiresOne(artifact.name(), required, groupID, hashesToText), "requiresOne")
			}
		}

		log("provides:")
		for _, provided := range artifact.provides(){
			provided = insertVariables(provided, artifact.name(), "", attributes, globals)
			log("\t",provided,md5hash(fmt.Sprint(provided)))
			core.addLine(provide(artifact.name(), provided, hashesToText), "provides")
		}

		/*
		log("Conditional provides:")
		*/
	}

	log("\n")
	log(attributes)
	//log(hashesToText)

	/* --- STORE FEATURES ---*/
	featureModelRoot := newAbstractFeature("")
	features := map[featureName]Feature{"":featureModelRoot}

	for _, f := range json["features"].([]any){
		name := newFeatureName(f, len(features))
		features[name] = newFeature(name, getArtifactsFromFeatureJSON(f), artifacts)
		featureModelRoot.children[name] = true
	}
	
	// ALGORITMO DI AIDE SUI TAG
	generateFeatureTree("", features)
	
	log("\n\n")
	for _, feature := range features{
		log(feature)
	}
	core.runProgram()
	printTree("", 0, features)
	//fmt.Println(core.getProgram())
	//prologQueryConsole(core, hashesToText)	
}