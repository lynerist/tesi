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
		generateFeatureTree("", state.features)
		outJson, _ := extractCytoscapeJSON(state)
		w.Write(outJson)
	})
}

func handleVariableUpdate(state *State){
	http.HandleFunc("/updateAttribute", func(w http.ResponseWriter, r *http.Request) {	
		name := r.FormValue("name")
		value := r.FormValue("value")
		feature	 := featureName(r.FormValue("feature"))

		if isGlobal, _ := strconv.ParseBool(r.URL.Query().Get("isglobal")); isGlobal{
			global := variableName(name[strings.IndexRune(name, GLOBALSIMBLE):])
			state.globals.elected[global] = value
			updatePossibleProvidersByGlobalChange(global, state)
		}else{
			artifact := artifactName(name[:strings.IndexRune(name, VARIABLESIMBLE)])
			variable := variableName(name[strings.IndexRune(name, VARIABLESIMBLE):])
			state.variables[artifact][feature][variable] = value
			updatePossibleProvidersByVariableChange(artifact, feature, state)
		}

		outJson, _ := extractCytoscapeJSON(state)
		w.Write(outJson)
	})
}
