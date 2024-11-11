package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func startLocalServer(){
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "interface/" + r.URL.Path[1:])
	})
	fmt.Println(http.ListenAndServe(":"+PORT, nil))
}

//Response with interface json
func handleJSONLoading(state *State) {
	http.HandleFunc("/loadjson", func(w http.ResponseWriter, r *http.Request) {	
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
		outJson, _ := checkDeadFeaturesANDextractInterfaceJSON(state)
		w.Write(outJson)
	})
}

//Response with full interface json
func handleVariableUpdate(state *State){
	http.HandleFunc("/updateAttribute", func(w http.ResponseWriter, r *http.Request) {	
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

		outJson, _ := checkDeadFeaturesANDextractInterfaceJSON(state)
		w.Write(outJson)
	})
}

//Response with active nodes list
func handleActivation(state *State){
	http.HandleFunc("/activation", func(w http.ResponseWriter, r *http.Request) {	
		feature	 := featureName(r.FormValue("feature"))
		if _, isDead := state.deadFeatures[feature]; !isDead{
			if _, isActive := state.activeFeatures[feature]; isActive{
				unactivateDown(feature, state)
			}else{
				activateUp(feature, state)
			}
		}
		w.Write(state.activeFeatures.jsonFormat())
	})
}

//Response with ???
func handleValidation(state *State){
	http.HandleFunc("/validation", func(w http.ResponseWriter, r *http.Request) {	
		w.Write([]byte("ciao"))
	})
}