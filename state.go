package main

type State struct {
	hashesToText map[string]string
	core prologCore
	artifacts map[artifactName]Artifact
	attributes map[artifactName]map[featureName]map[variableName]any
	globals globalContext
	features map[featureName]Feature
}

func newState()(state State){
	state.reset()
	return state
}

func (state *State) reset(){
	state.hashesToText = make(map[string]string)
	state.artifacts = make(map[artifactName]Artifact)
	state.attributes = make(map[artifactName]map[featureName]map[variableName]any)
	state.globals = newGlobalContext()
	state.features = map[featureName]Feature{"":newAbstractFeature("")}
}