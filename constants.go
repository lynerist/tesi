package main

const VERBOSE = false
const ALL 		string = "all"
const NOT 		string = "not"
const ANY 		string = "any"
const ONE 		string = "one"
const REQUIRES 	string = "requires"
const PROVIDES 	string = "provides"

const PORT = "60124"

type artifactName 	string
type featureName	string
type tagName		string
type variableName	string
type hash 			string
type variableValue	any
type declaration any

type set[T comparable] map[T]bool