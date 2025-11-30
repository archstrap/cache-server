package parser

import (
	"log"
	"strconv"
	"strings"
)

type InputParser struct {
	InputMessage string
}

func (p *InputParser) Parse() (string, error) {

	inputMessage := strings.TrimSpace(p.InputMessage)
	if inputMessage == "" || len(inputMessage) <= 0 {
		log.Fatalln("Hey there, received an empty command")
	}

	log.Println("input messsage", inputMessage)

	tokens := strings.Split(inputMessage, "\r\n")
	log.Println("tokens: ", tokens)

	commandLength, err := strconv.Atoi(tokens[1][1:])

	if commandLength == 0 || err != nil {
		return "OK", nil
	}

	command := strings.ToUpper(strings.TrimSpace(tokens[2]))

	log.Println("received command: ", command)

	switch command {
	case "PING":
		return "PONG", nil
	case "ECHO":

		// TODO: revisit on this to perfectly align
		data := "hi"

		return data, nil
	default:
		return "OK", nil

	}

}
