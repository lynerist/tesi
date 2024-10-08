package main

type State struct {
	hashesToText map[hash]string
	core prologCore
	artifacts map[artifactName]Artifact
	attributes map[artifactName]map[featureName]map[variableName]variableValue
	globals globalContext
	features map[featureName]Feature
	providers map[declaration]set[featureName]
}

func newState()(state State){
	state.reset()
	return state
}

func (state *State) reset(){
	state.hashesToText = make(map[hash]string)
	state.artifacts = make(map[artifactName]Artifact)
	state.attributes = make(map[artifactName]map[featureName]map[variableName]variableValue)
	state.globals = newGlobalContext()
	state.features = map[featureName]Feature{"":newAbstractFeature("")}
}