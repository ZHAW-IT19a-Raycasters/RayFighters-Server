package main

import (
	"fmt"

	"github.zhaw.ch/raycasters/server/utils"
)

func main() {
	// shutting of annoying error
	args, arparser, err := utils.ArgParse()

	if err != nil {
		arparser.Help("An error occured, while parsing the arguments.")
	}

	fmt.Printf("%s, %d", *args.Ip, *args.Port)

	// conf, err := utils.ConfigParse()
	// if err != nil {
	// 	panic("Parsing of the configuration failed")
	//
}
