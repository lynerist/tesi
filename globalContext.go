package main

type globalContext struct {
	proposed map[variableName]map[variableValue]int
	elected map[variableName]variableValue
	usedBy map[variableName]set[featureName]
	neededByArtifact map[artifactName]set[variableName]
}
func newGlobalContext()(newContext globalContext){
	newContext.proposed = make(map[variableName]map[variableValue]int)
	newContext.elected = make(map[variableName]variableValue)
	newContext.usedBy = make(map[variableName]set[featureName])
	newContext.neededByArtifact = make(map[artifactName]set[variableName])
	return
}
func (gr globalContext)put(name variableName, value variableValue, artifact artifactName){
	if len(gr.proposed[name])==0{
		gr.proposed[name] = make(map[variableValue]int)
	} 
	gr.proposed[name][value]++
	if gr.proposed[name][value] > gr.proposed[name][gr.elected[name]] || len(gr.proposed)==1{
		gr.elected[name]=value
	}
	if len(gr.neededByArtifact[artifact])==0{
		gr.neededByArtifact[artifact] = make(set[variableName])
	}
	gr.neededByArtifact[artifact].add(name)
}
func (gr globalContext)get(name variableName)variableValue{
	return gr.elected[name]
}
