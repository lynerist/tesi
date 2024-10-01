package main

import (
	"fmt"
	"strings"
)

type Feature struct {
	name featureName
	artifacts []artifactName
	tags map[tagName]bool
	children map[featureName]bool
	parent *featureName
}

func newFeatureName(feature any, id int)featureName{
	return featureName(fmt.Sprintf("%s::%d", feature.(map[string]any)["name"].(string), id))
}

func (f featureName) String()string{
	if f==""{
		return "Feature Model Root"
	}
	return string(f)
}

func newFeature(name featureName, composingArtifacts []any, artifacts map[artifactName]Artifact, parent featureName)Feature{
	feature := Feature{name, nil, make(map[tagName]bool), make(map[featureName]bool), &parent}

	for _, artifact := range composingArtifacts {
		feature.artifacts = append(feature.artifacts, artifactName(artifact.(string)))
		for _, tag := range artifacts[stringToAN(artifact)].tags(){
			feature.tags[tagName(tag.(string))] = true
		}
	}
	return feature
}

func newAbstractFeature(name featureName)Feature{
	return Feature{name, nil, nil, make(map[featureName]bool), new(featureName)}
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
