package main

import (
	"fmt"
)

type Feature struct {
	name featureName
	artifacts set[artifactName] 
	tags set[tagName]
	children set[featureName]
	parent *featureName
	requirements map[artifactName]Requirements
	provisions map[artifactName]set[declaration] 
}

func newFeatureName(feature any, id int)featureName{
	return featureName(fmt.Sprintf("%s::%d", feature.(map[string]any)["name"].(string), id))
}

func (f featureName) String()string{
	if f == ROOT {
		return "root"
	}
	return string(f)
}

func newFeature(name featureName, composingArtifacts []any, artifacts map[artifactName]Artifact, parent featureName)Feature{
	feature := Feature{name, make(set[artifactName]), make(set[tagName]), make(set[featureName]), 
		&parent, make(map[artifactName]Requirements), make(map[artifactName]set[declaration])}
	
	for _, artifact := range composingArtifacts {
		feature.artifacts.add(artifactName(artifact.(string)))
		feature.tags.add(artifacts[artifactName(artifact.(string))].tags)
	}
	return feature
}

func newAbstractFeature(name featureName)Feature{
	return Feature{name, nil, nil, make(set[featureName]), nil, nil, nil}
}

func (f Feature) String()string {
	var tags []tagName
	for tag := range f.tags{
		tags = append(tags, tag)
	}
	var children []featureName
	for child := range f.children{
		children = append(children, child)
	}

	var parent any = ROOT
	if f.parent != nil {
		parent = *f.parent
	}

	return fmt.Sprintf("'%s' --> artifacts:%v tags: %v children: %v parent: %s", f.name, f.artifacts, tags, children, parent)
}

func generateFeatureTree(root featureName, features map[featureName]Feature){
	tagCount := make(map[tagName]int)
	for child := range features[root].children{
		for tag := range features[child].tags{
			tagCount[tag]++
		}
	}

	for{
		var mostPresentTag tagName 
		for tag, count := range tagCount{
			if count > tagCount[mostPresentTag] && count>1{
				mostPresentTag = tag
			}
		}
		if mostPresentTag == ""{
			break
		}

		newTagNode := newAbstractFeature(featureName(fmt.Sprintf("%s::%d",mostPresentTag, len(features))))
		newTagNode.parent = &root
		for child := range features[root].children{
			if _, ok := features[child].tags[mostPresentTag]; ok{
				newTagNode.children.add(child)
			}
		}
		for child := range newTagNode.children{
			for tag := range features[child].tags{
				tagCount[tag]--
			}
			*features[child].parent = newTagNode.name
			delete(features[child].tags, mostPresentTag)
			delete(features[root].children, child)
		}
		features[root].children.add(newTagNode.name)
		features[newTagNode.name] = newTagNode
	}
	for child := range features[root].children{
		generateFeatureTree(child, features)
	}
}

func storeFeatures(json map[string]any, state *State){
	for _, f := range json["features"].([]any){
		feature := newFeature(newFeatureName(f, len(state.features)), getArtifactsFromFeatureJSON(f), state.artifacts, ROOT)
		for thisArtifactName := range feature.artifacts{
			artifact := state.artifacts[thisArtifactName]
			feature.requirements[thisArtifactName]= newRequirements()

			/* --- STORE FEATURES USING GLOBALS --- */
			for global := range state.artifacts[thisArtifactName].globalsDefault{
				if state.globals.usedBy[global] == nil {
					state.globals.usedBy[global] = make(set[featureName])
				}
				state.globals.usedBy[global].add(feature.name)
			}

			/* --- STORE ARTIFACT VARIABLES --- */
			for variable, value := range state.artifacts[thisArtifactName].variablesDefault{
				state.variables[thisArtifactName][feature.name] = make(map[attributeName]attributeValue) 
				state.variables[thisArtifactName][feature.name][variable] = value
			}

			/* --- STORE REQUIRED DECLARATIONS --- */
			feature.requirements[thisArtifactName].ALL.add(artifact.requiresALL())
			feature.requirements[thisArtifactName].NOT.add(artifact.requiresNOT())

			for _, group := range artifact.requiresANY(){
				declarations := make(set[declaration])
				declarations.add(group)
				*feature.requirements[thisArtifactName].ANY = append(*feature.requirements[thisArtifactName].ANY, declarations)
			}

			for _, group := range artifact.requiresONE(){
				declarations := make(set[declaration])
				declarations.add(group)
				*feature.requirements[thisArtifactName].ONE = append(*feature.requirements[thisArtifactName].ONE, declarations)
			}

			/* --- STORE PROVIDED DECLARATIONS --- */
			feature.provisions[thisArtifactName] = make(set[declaration])
			
			for provided := range artifact.provides{
				//I store the untouched declaration with variables and globals
				feature.provisions[thisArtifactName].add(provided)

				//I insert the actual values of the attributes and store the provided atom
				atom := insertAttributes(provided, thisArtifactName, feature.name, state)
				if state.possibleProviders[atom] == nil {
					state.possibleProviders[atom] = make(set[featureName])
				}
				state.possibleProviders[atom].add(feature.name)
			}
		}
		state.features[feature.name] = feature
		state.features[ROOT].children.add(feature.name)
	}
}

func (feature Feature)getRequirements(state *State) Requirements{
	requirements := newRequirements()
	for artifact, req := range feature.requirements{
		for atom := range req.ALL{
			requirements.ALL.add(insertAttributes(atom, artifact, feature.name, state))
		}

		for atom := range req.NOT{
			requirements.NOT.add(insertAttributes(atom, artifact, feature.name, state))
		}

		for _, group := range *req.ANY{
			declarations := make(set[declaration])
			for atom := range group {		
				declarations.add(insertAttributes(atom, artifact, feature.name, state))
			}
			*requirements.ANY = append(*requirements.ANY, declarations)
		}

		for _, group := range *req.ONE{
			declarations := make(set[declaration])
			for atom := range group {		
				declarations.add(insertAttributes(atom, artifact, feature.name, state))
			}
			*requirements.ONE = append(*requirements.ONE, declarations)
		}
	}
	return requirements
}

//return whenever this feature is actually providing the given declaration. (It may not provide it if its variables are changed)
func (feature Feature) isProviding(atom declaration, state *State)bool{
	for artifact, provideds := range feature.provisions{
		for provided := range provideds{
			if insertAttributes(provided, artifact, feature.name, state) == atom{
				return true
			}
		}
	}
	return false
}

func newRequirements()Requirements{
	return Requirements{make(set[declaration]), make(set[declaration]),new([]set[declaration]),new([]set[declaration])}
}

func activateUp(feature featureName, state *State){
	//TODO HANDLE DEAD FEATURES SOMEHOW
	state.activeFeatures.add(feature)
	if parent := state.features[feature].parent; parent != nil {
		activateUp(*parent, state)
	}
}

func unactivateDown(feature featureName, state *State) {
	state.activeFeatures.remove(feature)
	for child := range state.features[feature].children {
		unactivateDown(child, state)
	}
}