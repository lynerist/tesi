package main

type globalContext struct {
	proposed map[variableName]map[variableValue]int
	elected map[variableName]variableValue
	needs map[artifactName]set[variableName]
}
func newGlobalContext()(newContext globalContext){
	newContext.proposed = make(map[variableName]map[variableValue]int)
	newContext.elected = make(map[variableName]variableValue)
	newContext.needs = make(map[artifactName]set[variableName])
	return
}
func (gr globalContext)put(name variableName, value any, artifact artifactName){
	if len(gr.proposed[name])==0{
		gr.proposed[name] = make(map[variableValue]int)
	} 
	gr.proposed[name][value]++
	if gr.proposed[name][value] > gr.proposed[name][gr.elected[name]] || len(gr.proposed)==1{
		gr.elected[name]=value
	}
	if len(gr.needs[artifact])==0{
		gr.needs[artifact] = make(set[variableName])
	}
	gr.needs[artifact][name]=true
}
func (gr globalContext)get(name variableName)any{
	return gr.elected[name]
}
