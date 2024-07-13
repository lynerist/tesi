package main

import (
	"fmt"
	"github.com/ichiban/prolog"
)

func main(){
	json := readJSON("grammar.JSON")
	hashes 	:= make(map[string]string)
	dependencies := prolog.New(nil,nil)

	for _, production := range json{
		fmt.Println(production["name"])
		fmt.Println("requires:")
		if _, ok := production["requires"]; !ok {production["requires"]=[]any{}}
		if _, ok := production["provides"]; !ok {production["provides"]=[]any{}}
		if _, ok := production["conditionalProvides"]; !ok {production["conditionalProvides"]=[]any{}}

		for _, required := range production["requires"].([]any){
			requiredHash := md5hash(fmt.Sprint(require))
			fmt.Println("\t",required,requiredHash)
			hashes[requiredHash] = fmt.Sprint(required)
			dependencies.Exec(require(,))
		}
		fmt.Println("provides:")
		for _, provided := range production["provides"].([]any){
			fmt.Println("\t",provided,md5hash(fmt.Sprint(provided)))
			hashes[md5hash(fmt.Sprint(provided))] = fmt.Sprint(provided)

		}

		fmt.Println("Conditional provides:")
		for _, rule := range production["conditionalProvides"].([]any){
			fmt.Println("\tif got:")
			for _, needed := range rule.(map[string]any)["needs"].([]any){
				fmt.Println("\t\t",needed,md5hash(fmt.Sprint(needed)))
				hashes[md5hash(fmt.Sprint(needed))] = fmt.Sprint(needed)

			}
			fmt.Println("\tthen:")
			for _, provided := range rule.(map[string]any)["provides"].([]any){
				fmt.Println("\t\t",provided,md5hash(fmt.Sprint(provided)))
				hashes[md5hash(fmt.Sprint(provided))] = fmt.Sprint(provided)
			}
		}
		fmt.Println()
	}

	fmt.Println(hashes)

	
}