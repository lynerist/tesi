package main

import (
	"fmt"
	"github.com/ichiban/prolog"
)

func main(){
	json := readJSON("json/grammar.JSON")
	hashesToText 	:= make(map[string]string)
	dependencies := prolog.New(nil,nil)

	for _, production := range json{
		log(production["name"])
		if _, ok := production["requires"]; !ok {production["requires"]=[]any{}}
		if _, ok := production["provides"]; !ok {production["provides"]=[]any{}}
		if _, ok := production["conditionalProvides"]; !ok {production["conditionalProvides"]=[]any{}}
		
		log("requires:")
		for _, required := range production["requires"].([]any){
			log("\t",required,md5hash(fmt.Sprint(required)))
			dependencies.Exec(require(production["name"],required,hashesToText))
		}

		log("provides:")
		for _, provided := range production["provides"].([]any){
			log("\t",provided,md5hash(fmt.Sprint(provided)))
			dependencies.Exec(provide(production["name"], provided, hashesToText))
		}

		log("Conditional provides:")
		for _, rule := range production["conditionalProvides"].([]any){
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
	}

	log(hashesToText)

	
}