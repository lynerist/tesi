package main

type globalContext struct {
	proposed map[variableName]map[any]int
	elected map[variableName]any
	needs map[artifactName]map[variableName]bool
}
func newGlobalContext()(newRegister globalContext){
	newRegister.proposed = make(map[variableName]map[any]int)
	newRegister.elected = make(map[variableName]any)
	newRegister.needs = make(map[artifactName]map[variableName]bool)
	return
}
func (gr globalContext)put(name variableName, value any, artifact artifactName){
	if len(gr.proposed[name])==0{
		gr.proposed[name] = make(map[any]int)
	} 
	gr.proposed[name][value]++
	if gr.proposed[name][value] > gr.proposed[name][gr.elected[name]] || len(gr.proposed)==1{
		gr.elected[name]=value
	}
	if len(gr.needs[artifact])==0{
		gr.needs[artifact] = make(map[variableName]bool)
	}
	gr.needs[artifact][name]=true
}
func (gr globalContext)get(name variableName)any{
	return gr.elected[name]
}
