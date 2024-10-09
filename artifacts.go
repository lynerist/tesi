package main

type Artifact map[string]any 

func (a Artifact) name() artifactName{
	return artifactName(a["name"].(string))
}
func (a Artifact) requires(how string)[]any{
	return a[REQUIRES].(map[string]any)[how].([]any)
} 
func (a Artifact) provides()[]any{
	return a[PROVIDES].([]any)
} 
func (a Artifact) attributes()[]any{
	return a["attributes"].([]any)
} 
func (a Artifact) globals()[]any{
	return a["globals"].([]any)
} 
func (a Artifact) tags()[]any{
	return a["tags"].([]any)
}

func stringToAN(s any)artifactName{
	return artifactName(s.(string))
}

func (a Artifact) isVariadic()bool{
	return len(a.attributes()) + len(a.globals()) > 0
}

func storeArtifacts(json map[string]any, state *State){
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
}