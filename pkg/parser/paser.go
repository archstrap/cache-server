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
	TypeBulkString   RespType = '$'
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
		data, err := parseArray(reader)
		return &RespValue{
			Value:    data,
			DataType: TypeArray,
		}, err

	case TypeSimpleString:
		simpleString, err := parseSimpleString(reader)
		return &RespValue{
			Value:    simpleString,
			DataType: TypeSimpleString,
		}, err

	case TypeBulkString:
		bulkString, err := parseBulkString(reader)
		return &RespValue{
			Value:    bulkString,
			DataType: TypeBulkString,
		}, err

	case TypeInteger:
		integer, err := parseInteger(reader)
		return &RespValue{
			Value:    integer,
			DataType: TypeInteger,
		}, err
	default:
		value.DataType = TypeSimpleString
		value.Value = "HELLO"
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

func parseBulkString(reader *bufio.Reader) (string, error) {
	identifierByter, _ := reader.ReadByte()

	if RespType(identifierByter) != TypeBulkString {
		// TODO add error
	}

	lengthOfTheString, _ := reader.ReadString('\n')
	length, _ := strconv.Atoi(lengthOfTheString)

	data, _ := reader.ReadString('\n')
	data = strings.TrimSpace(data)

	if length != len(data) {
		// TODO
	}

	return data, nil

}

func parseArray(reader *bufio.Reader) ([]string, error) {

	noOfElements, _ := parseInteger(reader)
	elements := make([]string, noOfElements)

	for i := 0; i < noOfElements; i++ {
		data, _ := parseBulkString(reader)
		elements[i] = data
	}

	return elements, nil

}
