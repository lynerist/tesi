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
		fmt.Println(production["name"])
		if _, ok := production["requires"]; !ok {production["requires"]=[]any{}}
		if _, ok := production["provides"]; !ok {production["provides"]=[]any{}}
		if _, ok := production["conditionalProvides"]; !ok {production["conditionalProvides"]=[]any{}}
		
		fmt.Println("requires:")
		for _, required := range production["requires"].([]any){
			fmt.Println("\t",required,md5hash(fmt.Sprint(required)))
			dependencies.Exec(require(production["name"],required,hashesToText))
		}

		fmt.Println("provides:")
		for _, provided := range production["provides"].([]any){
			fmt.Println("\t",provided,md5hash(fmt.Sprint(provided)))
			hashesToText[md5hash(fmt.Sprint(provided))] = fmt.Sprint(provided)

		}

		fmt.Println("Conditional provides:")
		for _, rule := range production["conditionalProvides"].([]any){
			fmt.Println("\tif got:")
			for _, needed := range rule.(map[string]any)["needs"].([]any){
				fmt.Println("\t\t",needed,md5hash(fmt.Sprint(needed)))
				hashesToText[md5hash(fmt.Sprint(needed))] = fmt.Sprint(needed)

			}
			fmt.Println("\tthen:")
			for _, provided := range rule.(map[string]any)["provides"].([]any){
				fmt.Println("\t\t",provided,md5hash(fmt.Sprint(provided)))
				hashesToText[md5hash(fmt.Sprint(provided))] = fmt.Sprint(provided)
			}
		}
		fmt.Println()
	}

	fmt.Println(hashesToText)

	
}