package analyzer

import (
	"fmt"
	"regexp"
	"strings"
)

//TokenType is a string that represents an specifc pattern for a lexeme
type TokenType string

/*Tokens used for lexical analysis*/
const (
	ID           TokenType = "id"
	Number       TokenType = "number"
	String       TokenType = "string"
	Path         TokenType = "path"
	ContinueLine TokenType = "continue_line"
	NewLine      TokenType = "new_line"
	SMinus       TokenType = "minus"
	SArrow       TokenType = "arrow"
	RExec        TokenType = "reserved_exec"
	RPath        TokenType = "reserved_path"
	LexError     TokenType = "lex_error"
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
			} else if getBoolMatch("/", c) { // multi-lines and path
				state = 4
				auxlex += c
			} else if getBoolMatch("\n", c) { // new line
				state = 8
				auxlex += c
			} else {
				if !getBoolMatch("\\s", c) {
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
			} else {
				if auxlex == "exec" {
					addToken(RExec)
				} else if auxlex == "path" {
					addToken(RPath)
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
		// "/" was received
		case 4:
			if getBoolMatch("*", c) {
				auxlex += c
				state = 6
			} else if getBoolMatch("[[:word:]]", c) {
				auxlex += c
				state = 7
			}
			break
		// string validation
		case 5:
			if strings.Index(auxlex, "\"") == 0 && strings.LastIndex(auxlex, "\"") == len(command)-1 {
				addToken(String)
			} else {
				addToken(LexError)
			}
			i--
			break
		// "/*" was received
		case 6:
			if getBoolMatch("\n", c) {
				auxlex += c
				addToken(ContinueLine)
			}
			break
		// "/any_letter" was received (potential path)
		case 7:
			if getBoolMatch("[[:word:]/]", c) {
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
		}
	}
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
