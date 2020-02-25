package main

import (
	_ "encoding/json"
	"flag"
	_ "path/filepath"
	"strings"
)

func main() {
	mongoClient := mongoConnect()
	var urlPtr = flag.String("u", "https://example.com", "URL to play")
	var startPtr = flag.Int("s", 0, "Start from")
	flag.Parse()
	if strings.HasPrefix(*urlPtr, "https://www.anilibria.tv") {
		var code = strings.Split(strings.Split(*urlPtr, "/")[4], ".")[0]
		//fmt.Println(code)
		anilibria(code, *startPtr, mongoClient)
	} else if strings.HasPrefix(*urlPtr, "https://animevost.org") {
		var code = strings.Split(strings.Split(*urlPtr, "/")[5], "-")[0]
		vost(code, *startPtr, mongoClient)
	} else if strings.HasPrefix(*urlPtr, "https://smotret-anime.online/") {
		smotr(mongoClient)
	}
}
