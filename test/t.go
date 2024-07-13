package main

import "fmt"

func main(){
	m := make(map[int]int)

	f(m)

	fmt.Println(m)
}

func f(m map[int]int){
	m[3]=4
	m[2]=4
	m[1]=4
	m[4]=4
	m[5]=4
	m[6]=4
	m[7]=4
	m[8]=4
	m[9]=4
	m[10]=4
	m[15]=4
	m[145]=4
	m[64]=4
	m[73]=4
	m[86]=4
	m[945]=4
	m[103]=4
	m[112]=4
	m[142]=4
}