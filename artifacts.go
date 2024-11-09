package main

import "fmt"

type Artifact struct{
	name artifactName
	requirements map[string]any //map[string]<set[declaration] | []set[declaration]>
	provides []any
	attributes []any
	globals []any
	tags []any
}

func newArtifact(artifactFields map[string]any)(artifact Artifact){
	artifact.name = artifactName(artifactFields["name"].(string))
	artifact.provides = artifactFields[PROVIDES].([]any)

	//Store requirements in a map from string to set[declaration] (for ALL and NOT) or []set[declaration] (for ANY and ONE)
	artifact.requirements = make(map[string]any)
	
	/* ALL */
	artifact.requirements[ALL] = make(set[declaration])
	for _, required := range artifactFields[REQUIRES].(map[string]any)[ALL].([]any) {
		artifact.requirements[ALL].(set[declaration]).add(declaration(fmt.Sprint(required)))
	}
	/* NOT */
	artifact.requirements[NOT] = make(set[declaration])
	for _, required := range artifactFields[REQUIRES].(map[string]any)[NOT].([]any) {
		artifact.requirements[NOT].(set[declaration]).add(declaration(fmt.Sprint(required)))
	}
	/* ANY */
	artifact.requirements[ANY] = make([]set[declaration], 0)
	for _, group := range artifactFields[REQUIRES].(map[string]any)[ANY].([]any) {
		declarations := make(set[declaration])
		for _, required := range group.([]any) {
			declarations.add(declaration(fmt.Sprint(required)))
		}
		artifact.requirements[ANY]=append(artifact.requirements[ANY].([]set[declaration]), declarations)
	}
	/* ONE */
	artifact.requirements[ONE] = make([]set[declaration], 0)
	for _, group := range artifactFields[REQUIRES].(map[string]any)[ONE].([]any) {
		declarations := make(set[declaration])
		for _, required := range group.([]any) {
			declarations.add(declaration(fmt.Sprint(required)))
		}
		artifact.requirements[ONE]=append(artifact.requirements[ONE].([]set[declaration]), declarations)
	}

	artifact.attributes = artifactFields["attributes"].([]any)
	artifact.globals = artifactFields["globals"].([]any)
	artifact.tags = artifactFields["tags"].([]any)
	
	return artifact
}

func (a Artifact) requiresALL()set[declaration]{
	return a.requirements[ALL].(set[declaration])
}
func (a Artifact) requiresNOT()set[declaration]{
	return a.requirements[NOT].(set[declaration])
}
func (a Artifact) requiresANY()[]set[declaration]{
	return a.requirements[ANY].([]set[declaration])
}
func (a Artifact) requiresONE()[]set[declaration]{
	return a.requirements[ONE].([]set[declaration])
} 

func stringToAN(s any)artifactName{
	return artifactName(s.(string))
}

func (a Artifact) isVariadic()bool{
	return len(a.attributes) + len(a.globals) > 0
}

func storeArtifacts(json map[string]any, state *State){ 
	for _, a := range json["artifacts"].([]any){
		artifact := newArtifact(a.(map[string]any))
		state.artifacts[artifact.name] = artifact

		/* --- STORE ATTRIBUTES DEFAULTS---*/ 		
		for _, attribute := range artifact.attributes{
			if name, ok := attribute.(map[string]any)["name"]; ok && []rune(name.(string))[0]==VARIABLESIMBLE{
				if value, ok := attribute.(map[string]any)["default"]; ok{
					state.variables[artifact.name] = make(map[featureName]map[variableName]variableValue)
					state.variables[artifact.name][""] = make(map[variableName]variableValue) 
					state.variables[artifact.name][""][variableName(name.(string))] = value
				}
			}
		}

		/* --- STORE GLOBALS --- */ 		
		for _, global := range artifact.globals{
			if name, ok := global.(map[string]any)["name"]; ok && []rune(name.(string))[0]==GLOBALSIMBLE{
				if value, ok := global.(map[string]any)["default"]; ok{
					state.globals.put(variableName(name.(string)), value, artifact.name)
				}
			}
		}

		log(artifact.name)	
	}
}