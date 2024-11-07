package main

type State struct {
	hashesToText 		map[hash]string
	core 				prologCore
	artifacts 			map[artifactName]Artifact
	variables 			map[artifactName]map[featureName]map[variableName]variableValue
	globals 			globalContext
	features 			map[featureName]Feature
	possibleProviders	map[declaration]set[featureName]
	activeFeatures		set[featureName]
	deadFeatures		set[featureName]
}

func newState()(state State){
	state.reset()
	return state
}

func (state *State) reset(){
	state.hashesToText 		= make(map[hash]string)
	state.artifacts 		= make(map[artifactName]Artifact)
	state.variables 		= make(map[artifactName]map[featureName]map[variableName]variableValue)
	state.globals 			= newGlobalContext()
	state.features 			= map[featureName]Feature{ROOT:newAbstractFeature(ROOT)}
	state.possibleProviders	= make(map[declaration]set[featureName])
	state.activeFeatures 	= make(set[featureName])
	state.deadFeatures		= make(set[featureName])
}