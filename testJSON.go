package main

import (
	"fmt"
)

func main(){

	/* --- READ JSON ---*/
	json := readJSON("json/switch.JSON")
	hashesToText := make(map[string]string)
	core := setupProlog()
	
	/* --- MAP ARTIFACTS TO FEATURES AND FEATURES TO ARTIFACTS ---*/

	artifactsInFeature := make(map[string][]string)
	featureWithArtifact := make(map[string][]string)
	for _, f := range json["features"].([]any){
		feature := FeatureArtifacts(f.(map[string]any))
		for _, artifact := range feature.artifacts(){
			artifactsInFeature[feature.name()]=append(artifactsInFeature[feature.name()], artifact.(string))
			featureWithArtifact[artifact.(string)]=append(featureWithArtifact[artifact.(string)], feature.name())
		}
	}
	
	/* --- STORE ARTIFACTS ---*/

	artifacts := make(map[string]Artifact)
	attributes := make(map[string]map[string]map[string]any)
	// attributes[artifactName][featureName][variableName]
	for _, a := range json["artifacts"].([]any){
		artifact := Artifact(a.(map[string]any))
		artifacts[artifact.name()] = artifact

		/* --- STORE ATTRIBUTES ---*/ // TODO GLOBALI
		//TODO LEN MAP A NOME NODI
		

		for _, attribute := range artifact.attributes(){
			if name, ok := attribute.(map[string]any)["name"]; ok{
				if value, ok := attribute.(map[string]any)["default"]; ok{
					attributes[artifact.name()] = make(map[string]map[string]any)
					for _, feature := range featureWithArtifact[artifact.name()]{
						attributes[artifact.name()][feature] = make(map[string]any) 
						attributes[artifact.name()][feature][name.(string)] = value
					}
					attributes[artifact.name()][""] = make(map[string]any) 
					attributes[artifact.name()][""][name.(string)] = value
				}
			}
		}

		log(artifact.name())
		
		for _, required := range artifact.requires("all"){
			log("requires all:")
			required = insertVariables(required, attributes[artifact.name()][""])
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresAll(artifact.name(),required,hashesToText), "requiresAll")
		}
		for _, required := range artifact.requires("not"){
			log("requires not:")
			required = insertVariables(required, attributes[artifact.name()][""])
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresNot(artifact.name(),required,hashesToText), "requiresNot")
		}

		for groupID, requiredAnyGroup := range artifact.requires("any"){
			log("\t","any of:",requiredAnyGroup)
			for _, required :=  range requiredAnyGroup.([]any){
				required = insertVariables(required, attributes[artifact.name()][""])
				log("\t\t",required,md5hash(fmt.Sprint(required)))
				core.addLine(requiresAny(artifact.name(), required, groupID, hashesToText), "requiresAny")
			}
		}

		for groupID, requiredAnyGroup := range artifact.requires("one"){
			log("\t","one of:",requiredAnyGroup)
			for _, required :=  range requiredAnyGroup.([]any){
				required = insertVariables(required, attributes[artifact.name()][""])
				log("\t\t",required,md5hash(fmt.Sprint(required)))
				core.addLine(requiresOne(artifact.name(), required, groupID, hashesToText), "requiresOne")
			}
		}

		log("provides:")
		for _, provided := range artifact.provides(){
			provided = insertVariables(provided, attributes[artifact.name()][""])
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
	features := map[string]Feature{"":featureModelRoot}

	for _, f := range json["features"].([]any){
		featureName := fmt.Sprintf("%s_%d", f.(map[string]any)["name"].(string), len(features))
		features[featureName] = newFeature(featureName, FeatureArtifacts(f.(map[string]any)), artifacts)
		featureModelRoot.children[featureName] = true
	}

	
	// ALGORITMO DI AIDE SUI TAG
	generateFeatureTree("", features)
	
	log("\n\n")
	for _, feature := range features{
		log(feature)
	}
	core.runProgram()

	//fmt.Println(core.getProgram())
	prologQueryConsole(core, hashesToText)	
}