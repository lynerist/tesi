package main

type Artifact map[string]any 

func (a Artifact) name() artifactName{
	return artifactName(a["name"].(string))
}
func (a Artifact) requires(how string)[]any{
	return a[REQUIRES].(map[string]any)[how].([]any)
} 
func (a Artifact) provides()[]any{
	return a[PROVIDES].([]any)
} 
func (a Artifact) attributes()[]any{
	return a["attributes"].([]any)
} 
func (a Artifact) globals()[]any{
	return a["globals"].([]any)
} 
func (a Artifact) tags()[]any{
	return a["tags"].([]any)
}

func stringToAN(s any)artifactName{
	return artifactName(s.(string))
}