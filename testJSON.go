package main

import (
	"fmt"
)

func main(){
	json := readJSON("json/switch.JSON")
	hashesToText := make(map[string]string)
	core := setupProlog()
	// attributes[artifactName][featureName][variableName]
	
	artifactsInFeature := make(map[string][]string)
	featureWithArtifact := make(map[string][]string)
	for _, f := range json["features"].([]any){
		feature := Feature(f.(map[string]any))
		for _, artifact := range feature.artifacts(){
			artifactsInFeature[feature.name()]=append(artifactsInFeature[feature.name()], artifact.(string))
			featureWithArtifact[artifact.(string)]=append(featureWithArtifact[artifact.(string)], feature.name())
		}
	}
	
	attributes := make(map[string]map[string]map[string]any)
	for _, a := range json["artifacts"].([]any){
		artifact := Artifact(a.(map[string]any))
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
		for _, rule := range artifact["conditionalProvides"].([]any){
			log("\tif got:")
			for _, needed := range rule.(map[string]any)["needs"].([]any){
				log("\t\t",needed,md5hash(fmt.Sprint(needed)))
				hashesToText[md5hash(fmt.Sprint(needed))] = fmt.Sprint(needed)

			}
			log("\tthen:")
			for _, provided := range rule.(map[string]any)["provides"].([]any){
				log("\t\t",provided,md5hash(fmt.Sprint(provided)))
				hashesToText[md5hash(fmt.Sprint(provided))] = fmt.Sprint(provided)
			}
		}
		log("")
		*/
	}

	log("\n")
	log(attributes)
	//log(hashesToText)

	core.runProgram()

	//fmt.Println(core.getProgram())
	prologQueryConsole(core, hashesToText)	
}