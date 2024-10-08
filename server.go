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
		for _, a := range json["artifacts"].([]any){
			artifact := Artifact(a.(map[string]any))
			state.artifacts[artifact.name()] = artifact
	
			/* --- STORE ATTRIBUTES DEFAULTS---*/ 		
			for _, attribute := range artifact.attributes(){
				if name, ok := attribute.(map[string]any)["name"]; ok && []rune(name.(string))[0]=='$'{
					if value, ok := attribute.(map[string]any)["default"]; ok{
						state.attributes[artifact.name()] = make(map[featureName]map[variableName]variableValue)
						state.attributes[artifact.name()][""] = make(map[variableName]variableValue) 
						state.attributes[artifact.name()][""][variableName(name.(string))] = value
					}
				}
			}
	
			/* --- STORE GLOBALS --- */ 		
			for _, global := range artifact.globals(){
				if name, ok := global.(map[string]any)["name"]; ok && []rune(name.(string))[0]=='@'{
					if value, ok := global.(map[string]any)["default"]; ok{
						state.globals.put(variableName(name.(string)), value, artifact.name())
					}
				}
			}
	
			log(artifact.name())	
		}

		/* --- STORE FEATURES --- */
		for _, f := range json["features"].([]any){
			feature := newFeature(newFeatureName(f, len(state.features)), getArtifactsFromFeatureJSON(f), state.artifacts, "")
			state.features[feature.name] = feature
			state.features[""].children[feature.name] = true //all features are root's children
			for _, artifact := range feature.artifacts{
				for variable, value := range state.attributes[artifact][""]{
					state.attributes[artifact][feature.name] = make(map[variableName]variableValue) 
					state.attributes[artifact][feature.name][variable] = value
				}
			}
		}

		/* --- FEATURE TREE GENERATION --- */
		generateFeatureTree("", state.features)
		outJson, _ := cytoscapeJSON(state.features)
		w.Write(outJson)
	})
}
