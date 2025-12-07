package model

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
	Command  string
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
