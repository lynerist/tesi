package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	j "encoding/json"

)

func startLocalServer(){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "interface/" + r.URL.Path[1:])
	})
	fmt.Println(http.ListenAndServe(":"+PORT, nil))
}

//Response with interface json
func handleJSONLoading(state *State) {
	http.HandleFunc(API_LOAD_JSON, func(w http.ResponseWriter, r *http.Request) {	
		/* --- RESET STATE --- */
		state.reset()

		/* --- READ JSON --- */
		file, _ , _ := r.FormFile("json")
		defer file.Close()
		json := parseJSON(file)

		/* --- STORE ARTIFACTS --- */
		storeArtifacts(json, state)

		/* --- STORE FEATURES --- */
		storeFeatures(json, state)

		/* --- FEATURE TREE GENERATION --- */
		generateFeatureTree(ROOT, state.features)
		outJson, _ := checkDeadFeaturesANDExtractInterfaceJSON(state)
		w.Write(outJson)
		logConfiguration(state)
	})
}

//Response with interface json
func handleJSONRequest(state *State) {
	http.HandleFunc(API_JSON_REQUEST, func(w http.ResponseWriter, r *http.Request) {	
		outJson, _ := checkDeadFeaturesANDExtractInterfaceJSON(state)
		w.Write(outJson)
		logConfiguration(state)
	})
}

//Response with full interface json
func handleVariableUpdate(state *State){
	http.HandleFunc(API_UPDATE_ATTRIBUTE, func(w http.ResponseWriter, r *http.Request) {	
		name := r.FormValue("name")
		value := r.FormValue("value")
		feature	 := featureName(r.FormValue("feature"))

		if isGlobal, _ := strconv.ParseBool(r.URL.Query().Get("isglobal")); isGlobal{
			global := attributeName(name[strings.IndexRune(name, GLOBALSIMBLE):])
			state.globals.elected[global] = value
			updatePossibleProvidersByGlobalChange(global, state)
		}else{
			artifact := artifactName(name[:strings.IndexRune(name, VARIABLESIMBLE)])
			variable := attributeName(name[strings.IndexRune(name, VARIABLESIMBLE):])
			state.variables[artifact][feature][variable] = value
			updatePossibleProvidersByVariableChange(artifact, feature, state)
		}

		outJson, _ := checkDeadFeaturesANDExtractInterfaceJSON(state)
		w.Write(outJson)
		logConfiguration(state)
	})
}

//Response with active nodes list
func handleActivation(state *State){
	http.HandleFunc(API_SWITCH_ACTIVATION, func(w http.ResponseWriter, r *http.Request) {	
		feature	 := featureName(r.FormValue("feature"))
		if _, isDead := state.deadFeatures[feature]; !isDead{
			if _, isActive := state.activeFeatures[feature]; isActive{
				unactivateDown(feature, state)
			}else{
				activateUp(feature, state)
			}
		}
		w.Write(state.activeFeatures.jsonFormat())
		logConfiguration(state)
	})
}

//Response with a map from invalid features to not fullfilled requirements and a map from declarations(the ones in the previous map) to providers
func handleValidation(state *State){
	http.HandleFunc(API_VALIDATE, func(w http.ResponseWriter, r *http.Request) {	
		invalidFeatureRequirements := validate(state)
		providers := findProvidersForSelectedDeclarations(invalidFeatureRequirements, state)
		response,_ :=j.Marshal(map[string]any{"invalids":invalidFeatureRequirements, "providers":providers})
		w.Write(response)
	})
}

//Response with true if it changes the option
func handleVerboseValidationSwitch(){
	http.HandleFunc(API_SWITCH_VERBOSE_VALIDATION, func(w http.ResponseWriter, r *http.Request) {
		VERBOSEVALIDATION = !VERBOSEVALIDATION
		w.Write([]byte("true"))
	})
}

//Response with exported json [{name: "featureName", 
//									attributes: [{name:"attributeName",value:"attributeValue"}]}]
func handleExporting(state *State){
	http.HandleFunc(API_EXPORT_CONFIGURATION, func(w http.ResponseWriter, r *http.Request) {
		w.Write(exportConfiguration(state))
	})
}