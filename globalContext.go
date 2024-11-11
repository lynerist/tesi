package main

type globalContext struct {
	proposed map[attributeName]map[attributeValue]int
	elected map[attributeName]attributeValue
	usedBy map[attributeName]set[featureName]
}
func newGlobalContext()(newContext globalContext){
	newContext.proposed = make(map[attributeName]map[attributeValue]int)
	newContext.elected = make(map[attributeName]attributeValue)
	newContext.usedBy = make(map[attributeName]set[featureName])
	return
}
func (gr globalContext)put(name attributeName, value attributeValue){
	if gr.proposed[name] == nil{
		gr.proposed[name] = make(map[attributeValue]int)
	} 
	gr.proposed[name][value]++
	if gr.proposed[name][value] > gr.proposed[name][gr.elected[name]] || len(gr.proposed)==1{
		gr.elected[name]=value
	}
}
func (gr globalContext)get(name attributeName)attributeValue{
	return gr.elected[name]
}
