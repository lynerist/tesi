package main

import(
	"fmt"
	"os"
	"bufio"
	"sort"
	"strings"
	"github.com/ichiban/prolog"
)

type prologCore struct{
	interpreter *prolog.Interpreter
	code map[string]string
}

func (core *prologCore) addLine(line, where string){
	core.code[where] += "\n"+line
}

func (core *prologCore)getProgram()string{
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", 
						core.code["start"],
						core.code[ALL],
						core.code[NOT],
						core.code[ANY],
						core.code[ONE],
						core.code[PROVIDES])
}

func (core prologCore) runProgram(){
	err:=core.interpreter.Exec(core.getProgram())
	if err != nil{
		fmt.Println("Error in the program:", err)
	}
}

func (core *prologCore) reset(){
	core.code[ALL] = "requiresAll(foo,foo)."
	core.code[NOT] = "requiresNot(foo,foo)."
	core.code[ANY] = "requiresAny(foo,foo,0)."
	core.code[ONE] = "requiresOne(foo,foo,0)."
	core.code[PROVIDES] = fmt.Sprintf("provides(%s,root).",md5hash("root"))
}

func setupProlog() prologCore{
	file, _ := os.Open("core.pl")
	defer file.Close()
	sc := bufio.NewScanner(file)
	var program string
	for sc.Scan(){
		program += sc.Text() + "\n"
	}
	return prologCore{prolog.New(nil,nil), map[string]string{
											"start":program, 
											ALL:"requiresAll(foo,foo).",
											NOT:"requiresNot(foo,foo).",
											ANY:"requiresAny(foo,foo,0).",
											ONE:"requiresOne(foo,foo,0).",
											}}
}

func prologQueryConsole(state *State){
	var prologErrorsMeaning = map[string]string {
		"EOF":"Missing end of the query.",
	}

	sc := bufio.NewScanner(os.Stdin)
	for fmt.Print("?- "); sc.Scan(); fmt.Print("?- "){
		query := sc.Text()
		if query == "q" || query == "exit" {break}
		for hash, fullName := range state.hashesToFeature{
			query = strings.ReplaceAll(query, string(fullName), string(hash))
		}
		for hash, fullName := range state.hashesToDeclaration{
			query = strings.ReplaceAll(query, string(fullName), string(hash))
		}

		solutions, err := state.core.interpreter.Query(query)
		if err != nil{
			if meaning, ok := prologErrorsMeaning[fmt.Sprint(err)]; ok{
				fmt.Printf("Errore in '%s': %s\n\n", query, meaning)
			}else{
				fmt.Printf("Errore in '%s': %v\n\n", query, err)
			}
			continue
		}

		var output string
		for solutions.Next(){
			variables := make(map[string]any)
			solutions.Scan(variables)
			if len(variables) == 0{
				output += "true."
			}
			toPrint := make([]string,0,len(output))
			for k,v := range variables{
				hashedVar := hash("'"+fmt.Sprint(v)+"'")
				var answer any
				answer, ok := state.hashesToDeclaration[hashedVar]
				if !ok {
					answer, ok = state.hashesToFeature[hashedVar]
					if !ok{
						answer = fmt.Sprint(v)
					}
				}
				toPrint = append(toPrint, fmt.Sprintf("%s:%s\t",k,answer))
			}
			sort.Strings(toPrint)
			output += strings.Join(toPrint, "\t") + "\n"
		}
		
		if output == "" {
			fmt.Print("false.\n")
		}
		fmt.Println(output)
	}
}

func hashFeature(feature featureName, state *State)hash{
	hashed := md5hash(string(feature))
	state.hashesToFeature[hashed] = feature
	return hashed
}

func hashDeclaration(atom declaration, state *State)hash{
	hashed := md5hash(string(atom))
	state.hashesToDeclaration[hashed] = atom
	return hashed
}

func isFeatureValid(feature featureName, state *State)bool{
	query := fmt.Sprintf("valid(%s).", hashFeature(feature, state))
	solutions, err := state.core.interpreter.Query(query)
	if err != nil{
		fmt.Printf("Errore in '%s': %v\n\n", query, err)	
		return false
	}
	return solutions.Next()
}

func findMissingRequirements(feature featureName, state *State)Requirements{
	requirements := newRequirements()

	query := fmt.Sprintf("requiresAll(%s, What), \\+ exists(What).", hashFeature(feature, state))
	solutions, err := state.core.interpreter.Query(query)
	if err != nil{
		fmt.Printf("Errore in '%s': %v\n\n", query, err)	
		return Requirements{}
	}

	
	for solutions.Next(){
		variables := make(map[string]any)
		solutions.Scan(variables)
		for _,v := range variables{			
			requirements.ALL.add(unHash(hash(fmt.Sprintf("'%s'", v)), state).(declaration))
		}
	}

	return requirements
}

func validate(state *State){
	//Populate knowledge base
	state.core.reset()
	for feature := range state.activeFeatures{
		featureHash := hashFeature(feature, state)
		for artifact, requirements := range state.features[feature].requirements{
			for genericAtom := range requirements.ALL{
				atom := insertAttributes(genericAtom, artifact, feature, state)
				state.core.addLine(fmt.Sprintf("requiresAll(%s,%s).", featureHash, hashDeclaration(atom, state)), 
									ALL)
			}

			for genericAtom := range requirements.NOT{
				atom := insertAttributes(genericAtom, artifact, feature, state)
				state.core.addLine(fmt.Sprintf("requiresNot(%s,%s).",featureHash,hashDeclaration(atom, state)), 
									NOT)
			}

			for groupID, group := range *requirements.ANY{
				for genericAtom := range group{
					atom := insertAttributes(genericAtom, artifact, feature, state)
					state.core.addLine(fmt.Sprintf("requiresAny(%s,%s,%d).", featureHash, hashDeclaration(atom, state), groupID), 
									ANY)
				}
			}

			for groupID, group := range *requirements.ONE{
				for genericAtom := range group{
					atom := insertAttributes(genericAtom, artifact, feature, state)
					state.core.addLine(fmt.Sprintf("requiresOne(%s,%s,%d).", featureHash, hashDeclaration(atom, state), groupID), 
									ONE)
				}
			}
		}

		for artifact, provisions := range state.features[feature].provisions{
			for genericAtom := range provisions {
				atom := insertAttributes(genericAtom, artifact, feature, state)
				state.core.addLine(fmt.Sprintf("provides(%s,%s).", featureHash, hashDeclaration(atom, state)), 
									PROVIDES)
			}
		}
	}
	// Compile knowledge base
	state.core.runProgram()

	prologQueryConsole(state)

	//Check for feature validity
	invalids := make(map[featureName]Requirements)
	for feature := range state.activeFeatures{
		if !state.features[feature].isAbstract() && !isFeatureValid(feature, state){
			invalids[feature] = findMissingRequirements(feature, state)
		}
	}

	fmt.Println(invalids)
}