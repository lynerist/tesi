package main

import (
	"fmt"
)

func main(){
	json := readJSON("json/testAny.JSON")
	hashesToText := make(map[string]string)
	core := setupProlog()

	for _, artifact := range json{
		log(artifact["name"])
		fillMissingKeys(artifact)
		
		log("requires all:")

		for _, required := range artifact["requires"].(map[string]any)["all"].([]any){
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresAll(artifact["name"],required,hashesToText), "requiresAll")
		}
		for _, required := range artifact["requires"].(map[string]any)["not"].([]any){
			log("\t",required,md5hash(fmt.Sprint(required)))
			core.addLine(requiresNot(artifact["name"],required,hashesToText), "requiresNot")
		}

		for groupID, requiredAnyGroup := range artifact["requires"].(map[string]any)["any"].([]any){
			log("\t","one of:",requiredAnyGroup)
			for _, required :=  range requiredAnyGroup.([]any){
				log("\t\t",required,md5hash(fmt.Sprint(required)))
				core.addLine(requiresAny(artifact["name"], required, groupID, hashesToText), "requiresAny")
			}
		}

		log("provides:")
		for _, provided := range artifact["provides"].([]any){
			log("\t",provided,md5hash(fmt.Sprint(provided)))
			core.addLine(provide(artifact["name"], provided, hashesToText), "provides")
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

	log(hashesToText)

	core.runProgram()

	//fmt.Println(core.getProgram())
	prologQueryConsole(core, hashesToText)	
}