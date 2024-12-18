package main

import (
	"os"
	"fmt"
)

func main(){
	state := newState()

	for i:=1; i<len(os.Args); i++{
		switch os.Args[i]{
		case "-verbose":
			STANDARD_OUTPUT = true
		case "-file":
			if i == len(os.Args)-1{
				fmt.Println("You must give a file name after -file")
				break
			}
			file, err := os.Open(os.Args[i+1])
			if err != nil {
				fmt.Println("You must give a valid file name after -file")
				fmt.Println(err)
			}

			json := parseJSON(file)
			state.reset()
			/* --- STORE ARTIFACTS --- */
			storeArtifacts(json, &state)
			
			/* --- STORE FEATURES --- */
			storeFeatures(json, &state)

			/* --- FEATURE TREE GENERATION --- */
			generateFeatureTree(ROOT, state.features)	

			i++
		}
	}
	
	handleJSONLoading(&state)
	handleJSONRequest(&state)
	handleVariableUpdate(&state)
	handleActivation(&state)
	handleValidation(&state)
	handleVerboseValidationSwitch()
	handleExporting(&state)

	startLocalServer()
}