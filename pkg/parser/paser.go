package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/archstrap/cache-server/pkg/model"
)

func Parse(ioReader io.Reader) (*model.RespValue, error) {

	reader := bufio.NewReader(ioReader)
	typeByte, err := reader.ReadByte()

	if err != nil {
		log.Fatalf("Unable to read byte.")
	}

	value := &model.RespValue{}

	switch model.RespType(typeByte) {
	case model.TypeArray:
		data, err := parseArray(reader)
		return &model.RespValue{
			Value:    data,
			Command:  data[0],
			DataType: model.TypeArray,
		}, err

	case model.TypeSimpleString:
		simpleString, err := parseSimpleString(reader)
		return &model.RespValue{
			Value:    simpleString,
			DataType: model.TypeSimpleString,
		}, err

	case model.TypeBulkString:
		bulkString, err := parseBulkString(reader)
		return &model.RespValue{
			Value:    bulkString,
			DataType: model.TypeBulkString,
		}, err

	case model.TypeInteger:
		integer, err := parseInteger(reader)
		return &model.RespValue{
			Value:    integer,
			DataType: model.TypeInteger,
		}, err
	default:
		value.DataType = model.TypeSimpleString
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

	if model.RespType(identifierByter) != model.TypeBulkString {
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
