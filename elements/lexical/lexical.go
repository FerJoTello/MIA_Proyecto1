package elements

import (
	"fmt"
	"regexp"
	"strings"
)

//TokenType is a string that represents an specifc pattern for a lexeme
type TokenType string

//Tokens used for lexical analysis
const (
	ID           TokenType = "id"
	FileName     TokenType = "file_name"
	Number       TokenType = "number"
	String       TokenType = "string"
	Path         TokenType = "path"
	ContinueLine TokenType = "continue_line"
	NewLine      TokenType = "new_line"
	SMinus       TokenType = "minus"
	SArrow       TokenType = "arrow"
	ResID        TokenType = "reserved_id"
	ResExec      TokenType = "reserved_exec"
	ResPath      TokenType = "reserved_path"
	ResSize      TokenType = "reserved_size"
	ResName      TokenType = "reserved_name"
	ResUnit      TokenType = "reserved_unit"
	ResType      TokenType = "reserved_type"
	ResAdd       TokenType = "reserved_add"
	ResDelete    TokenType = "reserved_delete"
	ResFit       TokenType = "reserved_fit"
	ResMkdisk    TokenType = "reserved_mkdisk"
	ResRmdisk    TokenType = "reserved_rmdisk"
	ResFdisk     TokenType = "reserved_fdisk"
	ResMount     TokenType = "reserved_mount"
	ResUnmount   TokenType = "reserved_unmount"
	ResPause     TokenType = "reserved_pause"
	ResTipo      TokenType = "reserved_tipo"
	LexError     TokenType = "lex_error"
	EndToken     TokenType = "end_token"
)

//Token struct used to save the token information (type and value)
type Token struct {
	Type  TokenType
	Value string
}

//used for the lexical analysis
var auxlex string
var state int
var tokenList []Token

//Analyze does a lexical analysis returning the identified tokens from a string
func Analyze(command string) []Token {
	tokenList = nil
	command = command + "#"
	state = 0
	for i := 0; i < len(command); i++ {
		c := string(command[i]) //c is the character at index "i" from the original command
		switch state {
		case 0:
			if getBoolMatch("[[:alpha:]]", c) { //any letter
				state = 1
				auxlex += c
			} else if getBoolMatch("-", c) { // "-"
				state = 2
				auxlex += c
			} else if getBoolMatch("\"", c) { // for strings
				state = 3
				auxlex += c
			} else if getBoolMatch("[\\\\]", c) { // multi-lines
				state = 4
				auxlex += c
			} else if getBoolMatch("/", c) { // and path
				state = 7
				auxlex += c
			} else if getBoolMatch("\n", c) { // new line
				state = 8
				auxlex += c
			} else if getBoolMatch("[\\d]", c) {
				state = 9
				auxlex += c
			} else {
				if getBoolMatch("#", c) && i == len(command)-1 {
					//fmt.Println("Analisis terminado")
					auxlex += "\n"
					addToken(NewLine)
				} else if !getBoolMatch("\\s", c) {
					fmt.Println("Error en la entrada, el caracter: \"" + c + "\" no es válido")
					auxlex += c
					addToken(LexError)
				}
			}
			break
		//reserved word, id
		case 1:
			if getBoolMatch("[[:word:]]", c) {
				auxlex += c
			} else if getBoolMatch("[.]", c) {
				state = 10
				auxlex += c
			} else {

				if strings.ToLower(auxlex) == "exec" {
					addToken(ResExec)
				} else if strings.ToLower(auxlex) == "path" {
					addToken(ResPath)
				} else if strings.ToLower(auxlex) == "size" {
					addToken(ResSize)
				} else if strings.ToLower(auxlex) == "name" {
					addToken(ResName)
				} else if strings.ToLower(auxlex) == "unit" {
					addToken(ResUnit)
				} else if strings.ToLower(auxlex) == "mkdisk" {
					addToken(ResMkdisk)
				} else if strings.ToLower(auxlex) == "rmdisk" {
					addToken(ResRmdisk)
				} else if strings.ToLower(auxlex) == "fdisk" {
					addToken(ResFdisk)
				} else if strings.ToLower(auxlex) == "type" {
					addToken(ResType)
				} else if strings.ToLower(auxlex) == "fit" {
					addToken(ResFit)
				} else if strings.ToLower(auxlex) == "delete" {
					addToken(ResDelete)
				} else if strings.ToLower(auxlex) == "add" {
					addToken(ResAdd)
				} else if strings.ToLower(auxlex) == "mount" {
					addToken(ResMount)
				} else if strings.ToLower(auxlex) == "unmount" {
					addToken(ResUnmount)
				} else if strings.ToLower(auxlex) == "id" {
					addToken(ResID)
				} else if strings.ToLower(auxlex) == "tipo" {
					addToken(ResTipo)
				} else if strings.ToLower(auxlex) == "add" {
					addToken(ResAdd)
				} else if strings.ToLower(auxlex) == "pause" {
					addToken(ResPause)
				} else {
					addToken(ID)
				}
				i-- //The character is not part of the expresion so it's re-evaluated
			}
			break
		// "-" was received
		case 2:
			if getBoolMatch(">", c) {
				auxlex += c
				addToken(SArrow) //"->" is a valid token
			} else {
				addToken(SMinus) //"-" is a valid token
				i--              //The character is not part of the expresion so it's re-evaluated
			}
			break
		// a quotation mark was received
		case 3:
			if !getBoolMatch("\"", c) && i != len(command)-1 {
				auxlex += c
			} else if getBoolMatch("\"", c) {
				auxlex += c
				state = 5
			} else if i == len(command)-1 {
				auxlex += c
				addToken(LexError)
			}
			break
		// "\" was received
		case 4:
			if getBoolMatch("[*]", c) {
				auxlex += c
				state = 6
			} else {
				addToken(LexError)
				i--
			}
			break
		// string validation
		case 5:
			if strings.Index(auxlex, "\"") == 0 && strings.LastIndex(auxlex, "\"") == len(auxlex)-1 {
				addToken(String)
			} else {
				addToken(LexError)
			}
			i--
			break
		// "\*" was received
		case 6:
			if getBoolMatch("\n", c) {
				auxlex += c
				addToken(ContinueLine)
			} else if !getBoolMatch("\\s", c) { //if it's a whitespace skips it, if not it's a lexical error
				addToken(LexError)
				i--
			}
			break
		// "/" was received (potential path)
		case 7:
			if getBoolMatch("[[:word:]/.]", c) {
				auxlex += c
			} else {
				addToken(Path)
				i--
			}
			break
		// "\n" was received
		case 8:
			if auxlex == "\n" {
				addToken(NewLine)
				i--
			}
			break
		case 9:
			if getBoolMatch("[\\d]", c) {
				auxlex += c
			} else {
				addToken(Number)
				i--
			}
			break
		// "word." was received
		case 10:
			if getBoolMatch("[[:word:]]", c) {
				auxlex += c
			} else {
				addToken(FileName)
				i--
			}
			break
		}
	}
	//printTokens()
	return tokenList
}

func addToken(typee TokenType) {
	t := Token{
		Type:  typee,
		Value: auxlex,
	}
	tokenList = append(tokenList, t)
	auxlex = ""
	state = 0
}

func getBoolMatch(reg string, compared string) bool {
	boolean, err := regexp.MatchString(reg, compared)
	if err != nil {
		return false
	}
	return boolean
}

func printTokens() {
	fmt.Println("|TokenType|Value|")
	for i := 0; i < len(tokenList); i++ {
		fmt.Print("|")
		fmt.Print(tokenList[i].Type)
		fmt.Print("|")
		fmt.Print(tokenList[i].Value)
		fmt.Println("|")
	}
}
