package model

import "fmt"

type RespType byte

const (
	TypeArray        RespType = '*'
	TypeInteger      RespType = ':'
	TypeSimpleString RespType = '+'
	TypeBulkString   RespType = '$'
	TypeError        RespType = '-'
)

type RespValue struct {
	DataType RespType
	Value    any
	Command  string
}

func (r *RespValue) String() string {
	return fmt.Sprintf("Type: %d, Command: %s, Value: %s", r.DataType, r.Command, r.Value)
}

func (r *RespValue) ArgsToStringSlice() []string {
	return r.Value.([]string)
}

func (r *RespValue) ToRespOutput() *RespOutput {
	return NewRespOutput(r.DataType, r.Value)
}

type RespOutput struct {
	RespType RespType
	Data     any
}

func NewRespOutput(respType RespType, data any) *RespOutput {
	return &RespOutput{
		RespType: respType,
		Data:     data,
	}
}

func NewUnknownCommandOutput(data any) *RespOutput {
	return NewRespOutput(TypeError, data)
}

func NewWrongNumberOfOutput(commandName string) *RespOutput {
	return NewRespOutput(TypeError, fmt.Sprintf("ERR wrong number of arguments for '%s' command", commandName))
}
