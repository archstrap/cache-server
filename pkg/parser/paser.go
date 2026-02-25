package parser

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/archstrap/cache-server/pkg/model"
)

type RespParser struct {
	reader *bufio.Reader
}

func NewRespParser(reader io.Reader) *RespParser {
	return &RespParser{
		reader: bufio.NewReader(reader),
	}
}

func (r *RespParser) Parse() (*model.RespValue, error) {

	reader := r.reader
	typeByte, err := reader.ReadByte()

	if err != nil {
		slog.Info("Unable to read byte from the input")
		return nil, err
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
	// identifierByter, _ := reader.ReadByte()
	//
	// if model.RespType(identifierByter) != model.TypeBulkString {
	// 	// TODO add error
	// }

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

	for i := range noOfElements {
		if identifierByter, err := reader.ReadByte(); err != nil || model.RespType(identifierByter) != model.TypeBulkString {
			return nil, fmt.Errorf("Unable to read bytes of type bulkString\n")
		}
		data, _ := parseBulkString(reader)
		elements[i] = data
	}

	return elements, nil

}

func (r *RespParser) ParseRDb() (string, error) {
	bufReader := r.reader
	typeByte, _ := bufReader.ReadByte()
	if typeByte != '$' {
		return "", fmt.Errorf("Unknown type received. %s", string(typeByte))
	}
	slog.Info("received type byte", slog.Any("data", string(typeByte)))

	length, _ := bufReader.ReadString('\n')
	size, _ := strconv.Atoi(strings.TrimSpace(length))
	buf := make([]byte, size)
	io.ReadFull(bufReader, buf)
	return string(buf[:size]), nil
}
