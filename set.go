package main

import (
	"fmt"
	j "encoding/json"
)

type set[T comparable] map[T]struct{} 
type valueOrSet[T comparable] interface{}

func (s set[T]) add (toAdd valueOrSet[T]){
	if s == nil {
		panic("insert in nil set")
	}
	switch toAdd := toAdd.(type) {
	case T:
		s[toAdd]=struct{}{}
	case set[T]:
		for k := range toAdd{
			s[k]=struct{}{}
		}
	}
}

func (s set[T]) remove (toDel valueOrSet[T]){
	if s == nil {
		return
	}
	switch toDel := toDel.(type) {
	case T:
		delete(s, toDel)
	case set[T]:
		for k := range toDel{
			delete(s,k)
		}
	}
}

func (s set[T]) String()string {
	if len(s)==0{
		return "empty"
	}
	out := "set{"
	for e := range s{
		out += fmt.Sprintf("%v ", e)
	}
	return out[:len(out)-1]+"}"
}

func (s set[T]) jsonFormat()[]byte{
	var list []T = make([]T,0, len(s))
	for element := range s{
		list = append(list, element)
	}
	json, _ := j.Marshal(list)
	return json
}
