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
	requirements map[artifactName]Requirements
	provisions map[artifactName]set[declaration] 
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
	feature := Feature{name, nil, make(set[tagName]), make(set[featureName]), &parent, make(map[artifactName]Requirements), make(map[artifactName]set[declaration])}
	
	for _, artifact := range composingArtifacts {
		feature.artifacts = append(feature.artifacts, artifactName(artifact.(string)))
		for _, tag := range artifacts[stringToAN(artifact)].tags(){
			feature.tags[tagName(tag.(string))] = true
		}
	}
	return feature
}

func newAbstractFeature(name featureName)Feature{
	return Feature{name, nil, nil, make(set[featureName]), new(featureName), nil, nil}
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
		feature := newFeature(newFeatureName(f, len(state.features)), getArtifactsFromFeatureJSON(f), state.artifacts, ROOT)
		for _, thisArtifactName := range feature.artifacts{
			artifact := state.artifacts[thisArtifactName]
			feature.requirements[thisArtifactName]= newRequirements()

			/* --- STORE ARTIFACT VARIABLES --- */
			for variable, value := range state.attributes[thisArtifactName][""]{
				state.attributes[thisArtifactName][feature.name] = make(map[variableName]variableValue) 
				state.attributes[thisArtifactName][feature.name][variable] = value
			}

			/* --- STORE REQUIRED DECLARATIONS --- */
			
			/* ALL */
			for _, required := range artifact.requires(ALL){
				feature.requirements[thisArtifactName].ALL[declaration(fmt.Sprint(required))] = true
			}

			/* NOT */
			for _, required := range artifact.requires(NOT){	
				feature.requirements[thisArtifactName].NOT[declaration(fmt.Sprint(required))] = true
			}

			/* ANY */
			
			for _, group := range artifact.requires(ANY){
				declarations := make(set[declaration])

				for _, required := range group.([]any){					
					declarations[declaration(fmt.Sprint(required))] = true
				}

				*feature.requirements[thisArtifactName].ANY = append(*feature.requirements[thisArtifactName].ANY, declarations)
			}

			/* ONE */
			for _, group := range artifact.requires(ONE){
				declarations := make(set[declaration])

				for _, required := range group.([]any){
					declarations[declaration(fmt.Sprint(required))] = true
				}
				*feature.requirements[thisArtifactName].ONE = append(*feature.requirements[thisArtifactName].ONE, declarations)
			}

			feature.provisions[thisArtifactName] = make(set[declaration])
			/* --- STORE PROVIDED DECLARATIONS --- */
			for _, provided := range artifact.provides(){
				feature.provisions[thisArtifactName][declaration(fmt.Sprint(provided))] = true

				atom := insertVariables(provided, thisArtifactName, feature.name, state)
				if _, ok := state.possibleProviders[atom]; !ok {
					state.possibleProviders[atom] = make(set[featureName])
				}
				state.possibleProviders[atom][feature.name] = true
			}
		}
		state.features[feature.name] = feature
		state.features[ROOT].children[feature.name] = true 
	}
}

func (feature Feature)getRequirements(state *State) Requirements{
	requirements := newRequirements()
	for artifact, req := range feature.requirements{
		for atom := range req.ALL{
			requirements.ALL[insertVariables(atom, artifact, feature.name, state)] = true
		}

		for atom := range req.NOT{
			requirements.NOT[insertVariables(atom, artifact, feature.name, state)] = true
		}

		for _, group := range *req.ANY{
			declarations := make(set[declaration])
			for atom := range group {		
				declarations[insertVariables(atom, artifact, feature.name, state)] = true
			}
			*requirements.ANY = append(*requirements.ANY, declarations)
		}

		for _, group := range *req.ONE{
			declarations := make(set[declaration])
			for atom := range group {		
				declarations[insertVariables(atom, artifact, feature.name, state)] = true
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
			if insertVariables(provided, artifact, feature.name, state) == atom{
				return true
			}
		}
	}
	return false
}

func newRequirements()Requirements{
	return Requirements{make(set[declaration]), make(set[declaration]),new([]set[declaration]),new([]set[declaration])}
}