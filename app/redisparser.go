package main

import (
	"fmt"
)

type InputParserObject struct {
	InputMessage string
}

func Parse(p *InputParserObject) (string, error) {

	data := p.InputMessage
	fmt.Println("data", data)

	return data, nil
}
