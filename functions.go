package main

import (
	j "encoding/json"
	"os"
	"io"
	"crypto/md5"
	"fmt"
)

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

func require(who, what string, hashes map[string]string)string{
	return fmt.Sprintf("require(%s,%s)", who,what)
}