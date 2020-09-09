package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	DiskManager "./elements/diskmanager"
	Lexical "./elements/lexical"
)

var tokens []Lexical.Token    //tokens list from the command analyzed
var actualToken Lexical.Token //actual token that is being analyzed
var controlIndex int          //the index of the actual token

func main() {
	DiskManager.MountedDisks = []DiskManager.MountedDisk{}
	scanner := bufio.NewScanner(os.Stdin)
	var command string
	for command != "fin" {
		fmt.Print("Ingresa un comando:\n")
		scanner.Scan()
		command = "exec -pAth->\"/home/fernando/Documentos/2020 2do Semestre/Archivos/Proyecto1/pruebas/Comandos.mia\"" //scanner.Text()
		tokens = Lexical.Analyze(command)
		//is necessary to check if the tokens correspond to a command
		checkCommands()
		break
	}
}

//to check commands
func checkCommands() {
	//fmt.Println("start checking commands...")
	controlIndex = 0
	actualToken = tokens[controlIndex] //esto se puede optimizar xd
	for controlIndex < len(tokens) {
		if compareActualToken(Lexical.ResExec) {
			nextToken()
			exec()
			break
		} else if compareActualToken(Lexical.ResMkdisk) {
			nextToken()
			mkdisk()
		} else if compareActualToken(Lexical.ResRmdisk) {
			nextToken()
			rmdisk()
		} else if compareActualToken(Lexical.ResFdisk) {
			nextToken()
			fdisk()
		} else if compareActualToken(Lexical.ResMount) {
			nextToken()
			mount()
		} else if compareActualToken(Lexical.ResUnmount) {
			nextToken()
			unmount()
		}
		//a command has executed
		nextToken()
		if controlIndex == -1 {
			fmt.Println("Ya estufas")
			break
		}
	}
}

//unmount
func unmount() {
	var ID string // variables used to hold the command values
	//after splitting the command is necessary to iterate over the array in order to save the requeried values to do an operation
	for !compareActualToken(Lexical.NewLine) {
		if compareActualToken(Lexical.SMinus) {
			//should be a parameter
			nextToken()
			if compareActualToken(Lexical.ResID) {
				nextToken()
				ID = getID()
			}
		} else {
			nextToken()
		}
	}
	DiskManager.Unmount(ID)
}

//mount
func mount() {
	var path, name string // variables used to hold the command values
	//after splitting the command is necessary to iterate over the array in order to save the requeried values to do an operation
	for !compareActualToken(Lexical.NewLine) {
		if compareActualToken(Lexical.SMinus) {
			//should be a parameter
			nextToken()
			if compareActualToken(Lexical.ResPath) {
				nextToken()
				path = getPathString()
			} else if compareActualToken(Lexical.ResName) {
				nextToken()
				name = getID()
			}
		} else {
			nextToken()
		}
	}
	DiskManager.Mount(path, name)
}

//fdisk
func fdisk() {
	var add, size int // variables used to hold the command values
	var path, name, unit, typee, fit, delete string
	//after splitting the command is necessary to iterate over the array in order to save the requeried values to do an operation
	for !compareActualToken(Lexical.NewLine) {
		if compareActualToken(Lexical.SMinus) {
			//should be a parameter
			nextToken()
			if compareActualToken(Lexical.ResSize) {
				nextToken()
				size = getNumber()
			} else if compareActualToken(Lexical.ResPath) {
				nextToken()
				path = getPathString()
			} else if compareActualToken(Lexical.ResName) {
				nextToken()
				name = getID()
			} else if compareActualToken(Lexical.ResUnit) {
				nextToken()
				unit = getID()
			} else if compareActualToken(Lexical.ResType) {
				nextToken()
				typee = getID()
			} else if compareActualToken(Lexical.ResFit) {
				nextToken()
				fit = getID()
			} else if compareActualToken(Lexical.ResDelete) {
				nextToken()
				delete = getID()
			}
		} else {
			nextToken()
		}
	}
	DiskManager.Fdisk(path, name, unit, typee, fit, delete, add, size)
}

//rmdisk
func rmdisk() {
	var path string
	if compareActualToken(Lexical.SMinus) {
		nextToken()
		//next token should be path since rmdisk just has one param
		if compareActualToken(Lexical.ResPath) {
			nextToken()
			path = getPathString()
		}
	}
	DiskManager.Rmdisk(path)
}

//command exec
//exec -path->/valid/file.mia
func exec() {
	var path string
	//exec was detected. the next token should be "-"
	if compareActualToken(Lexical.SMinus) {
		nextToken()
		//next token should be path since exec just has one param
		if compareActualToken(Lexical.ResPath) {
			nextToken()
			path = getPathString()
		}
	}
	if path != "" {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Println("Ocurrió un error al leer el archivo:\n" + err.Error())
			return
		}
		stringCont := string(content)
		if stringCont != "" {
			tokens = Lexical.Analyze(stringCont)
			checkCommands()
		}
	}
}

//command mkdisk
func mkdisk() {
	var size int // variables used to hold the command values
	var path, name, unit string
	//after splitting the command is necessary to iterate over the array in order to save the requeried values to do an operation
	for !compareActualToken(Lexical.NewLine) {
		if compareActualToken(Lexical.SMinus) {
			//should be a parameter
			nextToken()
			if compareActualToken(Lexical.ResSize) {
				nextToken()
				size = getNumber()
			} else if compareActualToken(Lexical.ResPath) {
				nextToken()
				path = getPathString()
			} else if compareActualToken(Lexical.ResName) {
				nextToken()
				name = getFileName()
			} else if compareActualToken(Lexical.ResUnit) {
				nextToken()
				unit = getID()
			}
		} else {
			nextToken()
		}
	}

	DiskManager.Mkdisk(path, name, unit, size)
}

//to check commands
func compareActualToken(comparedType Lexical.TokenType) bool {
	return actualToken.Type == comparedType
}

//to check commands
func nextToken() {
	controlIndex++
	if controlIndex < len(tokens) {
		actualToken = tokens[controlIndex] //no problem
	} else {
		controlIndex = -1 //there are no more elements on the array
	}
}

//params
func getNumber() int {
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.Number) {
			value, err := strconv.Atoi(actualToken.Value)
			if err == nil {
				nextToken()
				return value
			}
		}
	}
	return -1
}

func getPathString() string {
	var path string
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.Path) {
			path = actualToken.Value
			nextToken()
		} else if compareActualToken(Lexical.String) {
			path = strings.Trim(actualToken.Value, "\"")
			nextToken()
		}
	}
	return path
}

func getFileName() string {
	var name string
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.FileName) {
			name = actualToken.Value
			nextToken()
		}
	}
	return name
}

func getID() string {
	var unit string
	if compareActualToken(Lexical.SArrow) {
		nextToken()
		if compareActualToken(Lexical.ID) {
			unit = actualToken.Value
			nextToken()
		}
	}
	return unit
}

func pause() {
	fmt.Print("Ejecucion en pausa. Pulse Enter.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
