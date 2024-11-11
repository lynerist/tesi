package main

import "fmt"

type Artifact struct{
	name artifactName
	requirements map[string]any //map[string]<set[declaration] | []set[declaration]>
	provides set[declaration]
	variablesDefault map[attributeName]attributeValue
	globalsDefault map[attributeName]attributeValue
	tags set[tagName]
}

func newArtifact(artifactFields map[string]any)(artifact Artifact){
	artifact.name = artifactName(artifactFields["name"].(string))
	artifact.provides = make(set[declaration])

	//Store requirements in a map from string to set[declaration] (for ALL and NOT) or []set[declaration] (for ANY and ONE)
	artifact.requirements = map[string]any{	ALL:make(set[declaration]), 
											NOT:make(set[declaration]),
											ANY:make([]set[declaration], 0),
											ONE:make([]set[declaration], 0)}

	artifact.variablesDefault = make(map[attributeName]attributeValue)
	artifact.globalsDefault = make(map[attributeName]attributeValue)
	artifact.tags = make(set[tagName])
	
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

func storeArtifacts(json map[string]any, state *State){ 
	for _, a := range json["artifacts"].([]any){
		artifact := newArtifact(a.(map[string]any))
		artifactFields := a.(map[string]any)

		/* --- STORE REQUIREMENTS --- */

		/* ALL */
		for _, required := range artifactFields[REQUIRES].(map[string]any)[ALL].([]any) {
			artifact.requirements[ALL].(set[declaration]).add(declaration(fmt.Sprint(required)))
		}
		/* NOT */
		for _, required := range artifactFields[REQUIRES].(map[string]any)[NOT].([]any) {
			artifact.requirements[NOT].(set[declaration]).add(declaration(fmt.Sprint(required)))
		}
		/* ANY */
		for _, group := range artifactFields[REQUIRES].(map[string]any)[ANY].([]any) {
			declarations := make(set[declaration])
			for _, required := range group.([]any) {
				declarations.add(declaration(fmt.Sprint(required)))
			}
			artifact.requirements[ANY]=append(artifact.requirements[ANY].([]set[declaration]), declarations)
		}
		/* ONE */
		for _, group := range artifactFields[REQUIRES].(map[string]any)[ONE].([]any) {
			declarations := make(set[declaration])
			for _, required := range group.([]any) {
				declarations.add(declaration(fmt.Sprint(required)))
			}
			artifact.requirements[ONE]=append(artifact.requirements[ONE].([]set[declaration]), declarations)
		}

		/* --- STORE PROVIDED DECLARATIONS --- */
		for _, atom := range artifactFields[PROVIDES].([]any){
			artifact.provides.add(declaration(fmt.Sprint(atom)))
		}

		/* --- STORE VARIABLES DEFAULTS---*/ 		
		for _, variable := range artifactFields["variables"].([]any){
			if name, ok := variable.(map[string]any)["name"]; ok && []rune(name.(string))[0]==VARIABLESIMBLE{
				if value, ok := variable.(map[string]any)["default"]; ok{
					state.variables[artifact.name] = make(map[featureName]map[attributeName]attributeValue)
					artifact.variablesDefault[attributeName(name.(string))] = value
				}
			}
		}

		/* --- STORE GLOBALS --- */ 		
		for _, global := range artifactFields["globals"].([]any){	
			if name, ok := global.(map[string]any)["name"]; ok && []rune(name.(string))[0]==GLOBALSIMBLE{
				if value, ok := global.(map[string]any)["default"]; ok{
					artifact.globalsDefault[attributeName(name.(string))] = attributeValue(value)
					state.globals.put(attributeName(name.(string)), value)
				}
			}
		}

		/* --- STORE TAGS --- */
		for _, tag := range artifactFields["tags"].([]any){
			artifact.tags.add(tagName(fmt.Sprint(tag)))
		}

		state.artifacts[artifact.name] = artifact
	}
}