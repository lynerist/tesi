package main

//import "fmt"

func main(){
	
	state := newState()
	//core := setupProlog()

	handleJSONLoading(&state)

	startLocalServer()

}