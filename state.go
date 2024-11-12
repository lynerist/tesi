package main

type State struct {
	hashesToFeature		map[hash]featureName
	hashesToDeclaration	map[hash]declaration
	core 				prologCore
	artifacts 			map[artifactName]Artifact
	variables 			map[artifactName]map[featureName]map[attributeName]attributeValue
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
	state.hashesToFeature		= make(map[hash]featureName)
	state.hashesToDeclaration	= make(map[hash]declaration)
	state.core 					= setupProlog()
	state.artifacts		 		= make(map[artifactName]Artifact)
	state.variables 			= make(map[artifactName]map[featureName]map[attributeName]attributeValue)
	state.globals 				= newGlobalContext()
	state.features 				= map[featureName]Feature{ROOT:newAbstractFeature(ROOT)}
	state.possibleProviders		= make(map[declaration]set[featureName])
	state.activeFeatures 		= make(set[featureName])
	state.deadFeatures			= make(set[featureName])
}

func (state *State) isActive (feature featureName)bool{
	_, active := state.activeFeatures[feature]
	return active
}