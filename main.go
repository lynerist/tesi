package main

//import "fmt"

func main(){
	
	state := newState()
	//core := setupProlog()

	handleJSONLoading(&state)
	handleVariableUpdate(&state)
	handleActivation(&state)
	handleValidation(&state)

	startLocalServer()
}