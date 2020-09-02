package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	Lexical "./elements"
)

var tokens []Lexical.Token    //tokens list from the command analyzed
var actualToken Lexical.Token //actual token that is being analyzed
var controlIndex int          //the index of the actual token

type partition struct {
	Name   [20]byte
	Type   byte
	Fit    byte
	Status byte
	Start  uint64
}

type mbr struct {
	CreationDate  [20]byte
	DiskSignature uint64
	Partitions    [4]partition
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var command string
	for command != "fin" {
		fmt.Print("Ingresa un comando:\n")
		scanner.Scan()
		command = "MkDiSk -siZe->4 -pAth->\"/home/fernando/Documentos/2020 2do Semestre/Archivos/Proyecto1/pruebas\" -name->DiskSez.dsk -uniT->k" //scanner.Text()
		tokens = Lexical.Analyze(command)
		//is necessary to check if the tokens correspond to a command
		checkCommands()
		break
	}
}

//to check commands
func checkCommands() {
	fmt.Println("start checking commands...")
	controlIndex = 0
	actualToken = tokens[controlIndex]
	if compareActualToken(Lexical.RExec) {
		exec()
	} else if compareActualToken(Lexical.RMkdisk) {
		nextToken()
		mkdisk()
	}
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

//command exec
//exec -path->/valid/file.mia
func exec() {
	//var path string
	//exec was detected. the next token should be "-"
	if compareActualToken(Lexical.SMinus) {
		nextToken()
		//next token should be path since exec just has one param
		if compareActualToken(Lexical.Path) {
			nextToken()
			//path := rPath()
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
			if compareActualToken(Lexical.RSize) {
				nextToken()
				size = rSize()
			} else if compareActualToken(Lexical.RPath) {
				nextToken()
				path = rPath()
			} else if compareActualToken(Lexical.RName) {
				nextToken()
				name = rName()
			} else if compareActualToken(Lexical.RUnit) {
				nextToken()
				unit = rUnit()
			}
		} else {
			nextToken()
		}
	}
	//once the parameters are defined the disk can be created
	if size >= 0 && path != "" && name != "" {
		var sizeOnBytes int64
		if unit != "k" {
			//units on megabytes (1024*1024)
			sizeOnBytes = 1024 * 1024 * int64(size)
		} else {
			//units on kylobytes (1024)
			sizeOnBytes = 1024 * int64(size)
		}
		// partitions initialization
		var partitions [4]partition
		for i := 0; i < 4; i++ {
			copy(partitions[i].Name[:], "new_part")
			partitions[i].Status = 'F'
			partitions[i].Fit = 'x'
			partitions[i].Type = 'x'
			partitions[i].Start = 97
		}
		// disk init
		newMbr := mbr{}
		newMbr.Partitions = partitions
		newMbr.DiskSignature = uint64(time.Now().Unix())
		copy(newMbr.CreationDate[:], time.Now().String()[0:19])
		writeMbr(path, name, sizeOnBytes, newMbr)
	} else {
		fmt.Println("No es posible ejecutar el comando \"mkdisk\". Parametros no definidos")
	}

}

func writeMbr(path string, fileName string, sizeOnBytes int64, newMbr mbr) {
	createFolders(path)
	//the file is created with the provided name and path
	file, err := os.Create(path + "/" + fileName)
	//escribir en archivo
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	var cero int8 = 0
	s := &cero
	//Primer binario para iniciar escribiendo el 0 como inicio del archivo.
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, s)
	escribirBytes(file, binario.Bytes())
	//Posicionar en la ultima posicion-1 para cumplir con el tamano. Se escribe un 0
	file.Seek(sizeOnBytes-1, 0) // segundo parametro: 0, 1, 2.     0 -> Inicio, 1-> desde donde esta el puntero, 2 -> Del fin para atras
	//Segundo Binario para definir el tamanio del archivo
	var binario2 bytes.Buffer
	binary.Write(&binario2, binary.BigEndian, s)
	escribirBytes(file, binario2.Bytes())
	//Se escribe el struct en el inicio del archivo
	file.Seek(0, 0)
	s1 := &newMbr
	//binario para escribir en el archivo creado con el tamanio y con el struct definido
	var binario3 bytes.Buffer
	binary.Write(&binario3, binary.BigEndian, s1)
	escribirBytes(file, binario3.Bytes())
	fmt.Println("File created succesfully xd")
}

func escribirBytes(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

func createFolders(path string) {
	//checking if the path contains quotation marks
	if strings.Index(path, "\"") == 0 && strings.LastIndex(path, "\"") == len(path)-1 {
		path = strings.ReplaceAll(path, "\"", "")
	}
	//the path should exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		//if doesn't exist it's created
		mkdirErr := os.MkdirAll(path, os.ModePerm)
		if mkdirErr != nil {
			log.Fatal(mkdirErr)
		}
		//fmt.Println("Check path. Should exist by now")
	}
}

//params
func rSize() int {
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

func rPath() string {
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

func rName() string {
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

func rUnit() string {
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
