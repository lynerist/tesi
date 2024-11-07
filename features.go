package main

import (
	"fmt"
	"strings"
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
	if f=="" {
		return "Feature Model Root"
	}
	return string(f)
}

func newFeature(name featureName, composingArtifacts []any, artifacts map[artifactName]Artifact, parent featureName)Feature{
	feature := Feature{name, make(set[artifactName]), make(set[tagName]), make(set[featureName]), 
		&parent, make(map[artifactName]Requirements), make(map[artifactName]set[declaration])}
	
	for _, artifact := range composingArtifacts {
		feature.artifacts.add(artifactName(artifact.(string)))
		for _, tag := range artifacts[stringToAN(artifact)].tags(){
			feature.tags.add(tagName(tag.(string)))
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
		for thisArtifactName := range feature.artifacts{
			artifact := state.artifacts[thisArtifactName]
			feature.requirements[thisArtifactName]= newRequirements()

			/* --- STORE FEATURES USING GLOBALS --- */
			for global := range state.globals.neededByArtifact[thisArtifactName]{
				if state.globals.usedBy[global] == nil {
					state.globals.usedBy[global] = make(set[featureName])
				}
				state.globals.usedBy[global].add(feature.name)
			}

			/* --- STORE ARTIFACT VARIABLES --- */
			for variable, value := range state.variables[thisArtifactName][""]{
				state.variables[thisArtifactName][feature.name] = make(map[variableName]variableValue) 
				state.variables[thisArtifactName][feature.name][variable] = value
			}

			/* --- STORE REQUIRED DECLARATIONS --- */
			
			/* ALL */
			for _, required := range artifact.requires(ALL){
				feature.requirements[thisArtifactName].ALL.add(declaration(fmt.Sprint(required)))
			}

			/* NOT */
			for _, required := range artifact.requires(NOT){	
				feature.requirements[thisArtifactName].NOT.add(declaration(fmt.Sprint(required)))
			}

			/* ANY */
			
			for _, group := range artifact.requires(ANY){
				declarations := make(set[declaration])

				for _, required := range group.([]any){					
					declarations.add(declaration(fmt.Sprint(required)))
				}

				*feature.requirements[thisArtifactName].ANY = append(*feature.requirements[thisArtifactName].ANY, declarations)
			}

			/* ONE */
			for _, group := range artifact.requires(ONE){
				declarations := make(set[declaration])

				for _, required := range group.([]any){
					declarations.add(declaration(fmt.Sprint(required)))
				}
				*feature.requirements[thisArtifactName].ONE = append(*feature.requirements[thisArtifactName].ONE, declarations)
			}

			feature.provisions[thisArtifactName] = make(set[declaration])

			/* --- STORE PROVIDED DECLARATIONS --- */
			for _, provided := range artifact.provides(){
				feature.provisions[thisArtifactName].add(declaration(fmt.Sprint(provided)))

				atom := insertVariables(provided, thisArtifactName, feature.name, state)
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
			requirements.ALL.add(insertVariables(atom, artifact, feature.name, state))
		}

		for atom := range req.NOT{
			requirements.NOT.add(insertVariables(atom, artifact, feature.name, state))
		}

		for _, group := range *req.ANY{
			declarations := make(set[declaration])
			for atom := range group {		
				declarations.add(insertVariables(atom, artifact, feature.name, state))
			}
			*requirements.ANY = append(*requirements.ANY, declarations)
		}

		for _, group := range *req.ONE{
			declarations := make(set[declaration])
			for atom := range group {		
				declarations.add(insertVariables(atom, artifact, feature.name, state))
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