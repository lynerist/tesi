package main

const ALL 		string = "all"
const NOT 		string = "not"
const ANY 		string = "any"
const ONE 		string = "one"
const REQUIRES 	string = "requires"
const PROVIDES 	string = "provides"
const ROOT		featureName = "root"

const PORT = "60124"

type artifactName 	string
type featureName	string
type tagName		string
type attributeName	string
type hash 			string
type attributeValue	any
type declaration 	string

type Requirements struct {
	ALL set[declaration]
	NOT set[declaration]
	ANY *[]set[declaration]
	ONE *[]set[declaration]
}

var VARIABLESIMBLE 	rune = '$'
var GLOBALSIMBLE 	rune = '@'
var VERBOSEVALIDATION bool

//API CALLS
const API_LOAD_JSON = "/loadjson"
const API_UPDATE_ATTRIBUTE = "/updateAttribute"
const API_SWITCH_ACTIVATION = "/activation"
const API_VALIDATE = "/validation"
const API_SWITCH_VERBOSE_VALIDATION = "/verboseValidationSwitch"
const API_EXPORT_CONFIGURATION = "/exporting"
