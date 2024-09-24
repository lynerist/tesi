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
	program map[string]string
}

func (core *prologCore) addLine(s, where string){
	core.program[where] += "\n"+s
}

func (core *prologCore)getProgram()string{
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", 
						core.program["start"],
						core.program["requiresAll"],
						core.program["requiresNot"],
						core.program["requiresAny"],
						core.program["requiresOne"],
						core.program[PROVIDES])
}

func (core prologCore) runProgram(){
	err:=core.interpreter.Exec(core.getProgram())
	if err != nil{
		fmt.Println("Error in the program:", err)
	}
}

func setupProlog() prologCore{
	file, _ := os.Open("core.pl")
	sc := bufio.NewScanner(file)
	var program string
	for sc.Scan(){
		program += sc.Text() + "\n"
	}
	return prologCore{prolog.New(nil,nil), map[string]string{
											"start":program, 
											"requiresAll":"requiresAll(foo,foo).",
											"requiresNot":"requiresNot(foo,foo).",
											"requiresAny":"requiresAny(foo,foo,0).",
											"requiresOne":"requiresOne(foo,foo,0).",
											}}
}

func prologQueryConsole(core prologCore, hashes map[string]string){
	var prologErrorsMeaning = map[string]string {
		"EOF":"Missing end of the query.",
	}

	sc := bufio.NewScanner(os.Stdin)
	for fmt.Print("?- "); sc.Scan(); fmt.Print("?- "){
		query := sc.Text()
		for hash, fullName := range hashes{
			query = strings.ReplaceAll(query, fullName, hash)
		}

		solutions, err := core.interpreter.Query(query)
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
				answer, ok := hashes["'"+fmt.Sprint(v)+"'"]
				if !ok {answer = fmt.Sprint(v)}
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