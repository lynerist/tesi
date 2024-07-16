package main

import (
	j "encoding/json"
	"os"
	"io"
	"crypto/md5"
	"fmt"
)

const VERBOSE = true

func readJSON(fileName string)(json []map[string]any){
	jsonFile, _ := os.Open(fileName)
	jsonBin, _ := io.ReadAll(jsonFile)
	j.Unmarshal(jsonBin, &json)
	jsonFile.Close()
	return
}

func md5hash(s string)string{
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func require(who, what any, hashes map[string]string)string{
	requiredHash := md5hash(fmt.Sprint(what))
	hashes[requiredHash] = fmt.Sprint(what)
	return fmt.Sprintf("requires(%s,%s)", who,what)
}

func provide(who, what any, hashes map[string]string)string{
	providedHash := md5hash(fmt.Sprint(what))
	hashes[providedHash] = fmt.Sprint(what)
	return fmt.Sprintf("provides(%s,%s)", who,what)
}

func log(s ...any){
	if (VERBOSE){
		fmt.Println(s...)
	}
}