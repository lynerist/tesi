package main

import (
	"fmt"
)

func mainn(){

	/* --- READ JSON --- */

	jsonName := "workbanch"
	json := readJSON(fmt.Sprintf("json/%s.json", jsonName))
	hashesToText := make(map[string]string)
	core := setupProlog()
		
	/* --- STORE ARTIFACTS --- */
	artifacts := make(map[artifactName]Artifact)
	attributes := make(map[artifactName]map[featureName]map[variableName]any)
	globals := newGlobalContext()
	
	for _, a := range json["artifacts"].([]any){
		artifact := Artifact(a.(map[string]any))
		artifacts[artifact.name()] = artifact

		/* --- STORE ATTRIBUTES DEFAULTS---*/ 		
		for _, attribute := range artifact.attributes(){
			if name, ok := attribute.(map[string]any)["name"]; ok && []rune(name.(string))[0]=='$'{
				if value, ok := attribute.(map[string]any)["default"]; ok{
					attributes[artifact.name()] = make(map[featureName]map[variableName]any)
					attributes[artifact.name()][""] = make(map[variableName]any) 
					attributes[artifact.name()][""][variableName(name.(string))] = value
				}
			}
		}

		/* --- STORE GLOBALS --- */ 		
		for _, global := range artifact.globals(){
			if name, ok := global.(map[string]any)["name"]; ok && []rune(name.(string))[0]=='@'{
				if value, ok := global.(map[string]any)["default"]; ok{
					globals.put(variableName(name.(string)), value, artifact.name())
				}
			}
		}

		log(artifact.name())
		
		var state State
		for i, required := range artifact.requires(ALL){
			if i==0 {log("requires all:")}
			required = insertVariables(required, artifact.name(), "", &state)
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresAll(artifact.name(),required,hashesToText), "requiresAll")
		}
		for i, required := range artifact.requires(NOT){
			if i==0 {log("requires not:")}
			required = insertVariables(required, artifact.name(), "", &state)
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresNot(artifact.name(),required,hashesToText), "requiresNot")
		}

		for groupID, requiredAnyGroup := range artifact.requires(ANY){
			log("\t","any of:",requiredAnyGroup)
			for _, required :=  range requiredAnyGroup.([]any){
				required = insertVariables(required, artifact.name(), "", &state)
				log("\t\t",required,md5hash(fmt.Sprint(required)))
				core.addLine(requiresAny(artifact.name(), required, groupID, hashesToText), "requiresAny")
			}
		}

		for groupID, requiredAnyGroup := range artifact.requires(ONE){
			log("\t","one of:",requiredAnyGroup)
			for _, required :=  range requiredAnyGroup.([]any){
				required = insertVariables(required, artifact.name(), "", &state)
				log("\t\t",required,md5hash(fmt.Sprint(required)))
				core.addLine(requiresOne(artifact.name(), required, groupID, hashesToText), "requiresOne")
			}
		}

		log("provides:")
		for _, provided := range artifact.provides(){
			provided = insertVariables(provided, artifact.name(), "", &state)
			log("\t",provided,md5hash(fmt.Sprint(provided)))
			core.addLine(provides(artifact.name(), provided, hashesToText), "provides")
		}

		/*
		log("Conditional provides:")
		*/
	}

	log("\n")
	log(attributes)
	//log(hashesToText)

	/* --- STORE FEATURES --- */
	featureModelRoot := newAbstractFeature("")
	features := map[featureName]Feature{"":featureModelRoot}

	for _, f := range json["features"].([]any){
		feature := newFeature(newFeatureName(f, len(features)), getArtifactsFromFeatureJSON(f), artifacts, "")
		features[feature.name] = feature
		featureModelRoot.children[feature.name] = true
		for _, artifact := range feature.artifacts{
			for variable, value := range attributes[artifact][""]{
				attributes[artifact][feature.name] = make(map[variableName]any) 
				attributes[artifact][feature.name][variable] = value
			}
		}
	}
	
	/* --- FEATURE TREE GENERATION --- */
	generateFeatureTree("", features)
	
	log("\n\n")
	for _, feature := range features{
		log(feature)
	}
//	core.runProgram()
	printTree("", 0, features)
	//fmt.Println(core.getProgram())
	//prologQueryConsole(core, hashesToText)	

	exportFeatureModelJson("test/out.json", &State{})

	startLocalServer()
	
}