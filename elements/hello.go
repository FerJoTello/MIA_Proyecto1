package main

import (
	"bufio"
	Analyzer "elements/analyzer"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var command string
	for command != "fin" {
		fmt.Print("Ingresa un comando:\n")
		scanner.Scan()
		command = scanner.Text()
		tokens := Analyzer.Analyze(command)
		fmt.Println(len(tokens))
		for i := 0; i < len(tokens); i++ {
			fmt.Println(tokens[i].Type)
		}
	}
}

/*
func mkdisk(command string) {
	fmt.Println("Entró cagado de la risa")
	//var size, path, name, unit string // variables used to hold the command values
	var params []string
	params = strings.Split(command, " ") //each parameter is divided from the original command
	//after splitting the command is necessary to iterate over the array in order to save the requeried values to do an operation
	for i := 1; i < len(params); i++ {
		var param, value string
		param = strings.Split(params[i], "->")[0]
		value = strings.Split(params[i], "->")[1]
		switch strings.ToLower(param) {
		case "-size":
			//size = value
			break
		case "-path":
			//path = value
			break
		case "-name":
			//name = value
			break
		case "-unit":
			//unit = value
			break
		}
	}
	//once the parameters are defined the disk can be created

}
*/
