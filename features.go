package main

import (
	"fmt"
	"strings"
)

type Feature struct {
	name featureName
	artifacts []artifactName
	tags set[tagName]
	children set[featureName]
	parent *featureName
	requirements Requirements
	variadicRequirements Requirements
}

func newFeatureName(feature any, id int)featureName{
	return featureName(fmt.Sprintf("%s::%d", feature.(map[string]any)["name"].(string), id))
}

func (f featureName) String()string{
	if f=="" {
		return "Feature Model Root"
	}
	return string(f)
}

func newFeature(name featureName, composingArtifacts []any, artifacts map[artifactName]Artifact, parent featureName)Feature{
	feature := Feature{name, nil, make(set[tagName]), make(set[featureName]), &parent, 
		Requirements{make(set[declaration]),make(set[declaration]), nil, nil}, 
		Requirements{make(set[declaration]),make(set[declaration]), nil, nil}}
	
	for _, artifact := range composingArtifacts {
		feature.artifacts = append(feature.artifacts, artifactName(artifact.(string)))
		for _, tag := range artifacts[stringToAN(artifact)].tags(){
			feature.tags[tagName(tag.(string))] = true
		}
	}
	return feature
}

func newAbstractFeature(name featureName)Feature{
	return Feature{name, nil, nil, make(set[featureName]), new(featureName), Requirements{}, Requirements{}}
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

	return fmt.Sprintf("'%s' --> artifacts:%v tags: %v children: %v parent: %s", f.name, f.artifacts, tags, children, *f.parent)
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
		for child := range features[root].children{
			if features[child].tags[mostPresentTag]{
				newTagNode.children[child] = true
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
		features[root].children[newTagNode.name] = true
		features[newTagNode.name] = newTagNode
	}
	for child := range features[root].children{
		generateFeatureTree(child, features)
	}
}

func printTree(root featureName, indent int, features map[featureName]Feature){
	if len(features[root].children)==0{
		fmt.Printf("%s%s\n", strings.Repeat("\t", indent), root)
		return
	}
	fmt.Printf("%s%s -> [\n", strings.Repeat("\t", indent), root)
	for child := range features[root].children{
		printTree(child, indent+1, features)
	}
	fmt.Printf("%s]\n", strings.Repeat("\t", indent))
}

func storeFeatures(json map[string]any, state *State){
	for _, f := range json["features"].([]any){
		feature := newFeature(newFeatureName(f, len(state.features)), getArtifactsFromFeatureJSON(f), state.artifacts, "")
		for _, thisArtifactName := range feature.artifacts{
			artifact := state.artifacts[thisArtifactName]

			/* --- STORE ARTIFACT VARIABLES --- */
			for variable, value := range state.attributes[thisArtifactName][""]{
				state.attributes[thisArtifactName][feature.name] = make(map[variableName]variableValue) 
				state.attributes[thisArtifactName][feature.name][variable] = value
			}

			/* --- STORE REQUIRED DECLARATIONS --- */
			for _, required := range artifact.requires(ALL){
				atom := declaration(fmt.Sprint(required))
				if artifact.isVariadic(){
					atom = insertVariables(required, thisArtifactName, "", state)
					feature.variadicRequirements.ALL[atom] = true
				}else{
					feature.requirements.ALL[atom] = true
				}
			}

			for _, required := range artifact.requires(NOT){
				atom := declaration(fmt.Sprint(required))
				if artifact.isVariadic(){
					atom = insertVariables(required, thisArtifactName, "", state)
					feature.variadicRequirements.NOT[atom] = true
				}else{
					feature.requirements.NOT[atom] = true
				}
			}

			

			//TODO ANY ONE

			/* --- STORE PROVIDED DECLARATIONS --- */
			for _, provided := range artifact.provides(){
				atom := declaration(fmt.Sprint(provided))
				if artifact.isVariadic(){
					atom = insertVariables(provided, thisArtifactName, feature.name, state)
					if len(state.variadicProviders[atom])==0{state.variadicProviders[atom] = make(set[featureName])}
					state.variadicProviders[atom][feature.name]=true
				} else {
					if len(state.providers[atom])==0{state.providers[atom] = make(set[featureName])}
					state.providers[atom][feature.name]=true
				}
			}
		}
		state.features[feature.name] = feature
		state.features[""].children[feature.name] = true //all features are root's children
	}
}
