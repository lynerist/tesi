package main 

import (
	"net/http"
	"fmt"
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
		outJson, _ := cytoscapeJSON(state)
		w.Write(outJson)
	})
}
