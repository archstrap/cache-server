package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type RespType byte

const (
	TypeArray        RespType = '*'
	TypeInteger      RespType = ':'
	TypeSimpleString RespType = '+'
)

type RespValue struct {
	DataType RespType
	Value    any
}

func Parse(ioReader io.Reader) (*RespValue, error) {

	reader := bufio.NewReader(ioReader)
	typeByte, err := reader.ReadByte()

	if err != nil {
		log.Fatalf("Unable to read byte.")
	}

	value := &RespValue{}

	switch RespType(typeByte) {
	case TypeArray:
	case TypeSimpleString:
		simpleString, err := parseSimpleString(reader)
		return &RespValue{
			Value:    simpleString,
			DataType: TypeSimpleString,
		}, err

	case TypeInteger:
		integer, err := parseInteger(reader)
		return &RespValue{
			Value:    integer,
			DataType: TypeInteger,
		}, err
	}

	return value, nil

}

func parseInteger(reader *bufio.Reader) (int, error) {

	data, err := reader.ReadString('\n')
	if err != nil {
		return -1, err
	}

	data = strings.TrimSpace(data)
	integer, err := strconv.Atoi(data)

	fmt.Println(integer)

	return integer, err

}

func parseSimpleString(reader *bufio.Reader) (string, error) {
	data, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	data = strings.TrimSpace(data)
	return data, nil

}
