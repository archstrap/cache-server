package rdb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/archstrap/cache-server/internal/command"
	"github.com/archstrap/cache-server/internal/config"
	"github.com/archstrap/cache-server/pkg/model"
)

const (
	EOF          = 0xFF
	SELECTDB     = 0xFE
	EXPIRETIME   = 0xFD
	EXPIRETIMEMS = 0xFC
	RESIZEDB     = 0xFB
	AUX          = 0xFA
)

type RDBReader struct {
	file   *os.File
	reader *bufio.Reader
}

func LoadRdb() {
	rdbFilePath := fmt.Sprintf("%s/%s", config.Dir, config.Dbfilename)
	rdbReader := NewRDBReader(rdbFilePath)
	if rdbReader == nil {
		return
	}
	defer func() {
		rdbReader.Close()
	}()

	rdbReader.Read()
}

func NewRDBReader(path string) *RDBReader {

	file, err := os.Open(path)
	if err != nil {
		slog.Error("Unable to Set Up rdb file. Invalid RDB file provided. ", slog.Any("path", path))
		return nil
	}

	return &RDBReader{file: file, reader: bufio.NewReader(file)}
}

func (r *RDBReader) ReadHeader() (string, error) {

	header := make([]byte, 9)

	if _, err := io.ReadFull(r.reader, header); err != nil {
		return "", fmt.Errorf("Unable to Read Header from %s\n", r.file.Name())
	}

	magicString := string(header[:5])

	if "REDIS" != magicString {
		return "", fmt.Errorf("Not a valid REDIS rdb file. Check the contents of %s", r.file.Name())
	}

	version := string(header[5:])
	if "0011" != version {
		return "", fmt.Errorf("Not a valid REDIS version rdb file. Check the contents of %s", r.file.Name())
	}

	return fmt.Sprintf("%s%s", magicString, version), nil

}

func (r *RDBReader) Read() {

	if header, err := r.ReadHeader(); err != nil {
		slog.Error("", "ERR", err)
		return
	} else {
		slog.Info("", "RDB Header:", header)
	}
	for {
		opCode, err := r.ReadByte()
		if err == io.EOF {
			break
		}

		if opCode == EOF {
			slog.Info("EOF reached from the rdb  ", slog.Any("file", r.file.Name()))
			break
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		switch opCode {
		case AUX: // MetaData
			k, err := r.ReadString()
			if err != nil {
				fmt.Println(err)
				return
			}
			v, err := r.ReadString()
			if err != nil {
				fmt.Println(err)
				return
			}
			slog.Info("Read MetaData Section: ", slog.Any(k, v))
		case SELECTDB:
			_, err := r.ReadDb()
			if err != nil {
				fmt.Println(err)
				return
			}

		}

	}

}

func (r *RDBReader) ReadDb() (any, error) {

	dbIndex, err := r.ReadByte()
	if err != nil {
		return "", err
	}

	slog.Info("", "DB Index:", dbIndex)

	var hashTableSize, expiryTableSize byte

	expiryTimeStamp := ""
	timeStampType := ""
	for {

		typeByte, err := r.ReadByte()
		if err == io.EOF || typeByte == EOF {
			break
		}

		if err != nil {
			return "", err
		}

		switch typeByte {
		case 0xFB:
			hashTableSize, err = r.ReadByte()
			if err != nil {
				return "", err
			}
			slog.Info("Hash table ", slog.Any("size", int(hashTableSize)))
			expiryTableSize, err = r.ReadByte()
			if err != nil {
				return "", err
			}
			slog.Info("Expiry table ", slog.Any("size", int(expiryTableSize)))
		case 0xFD: // time stamp in seconds ( 4 byte unsigned integer) , LittleEndian

			buffer, err := r.ReadNBytes(4)
			if err != nil {
				return "", err
			}

			timeStamp := binary.LittleEndian.Uint32(buffer)
			expiryTimeStamp = fmt.Sprint(timeStamp)
			timeStampType = "PX"

		case 0xFC: // time stamp in miliseconds (8 byte unsigned long), LittleEndian
			buffer, err := r.ReadNBytes(8)
			if err != nil {
				return "", err
			}

			timeStamp := binary.LittleEndian.Uint64(buffer)
			expiryTimeStamp = fmt.Sprint(timeStamp)
			timeStampType = "EX"

		case 0:
			k, err := r.ReadString()
			if err != nil {
				return "", err
			}

			v, err := r.ReadString()
			if err != nil {
				return "", err
			}

			value := []string{"SET", k, v}

			if expiryTimeStamp != "" {
				value = append(value, timeStampType, expiryTimeStamp)
				timeStampType = ""
				expiryTimeStamp = ""
			}

			respValue := &model.RespValue{
				DataType: model.TypeArray,
				Value:    value,
				Command:  value[0],
			}
			command.SetCommandInstance.Process(respValue)
		}

	}

	return nil, nil

}

func (r *RDBReader) ReadString() (string, error) {

	var (
		prefix byte
		err    error
	)

	if prefix, err = r.ReadByte(); err != nil {
		return "", err
	}

	// Extracting first 2 bits
	// 0xC0 mask first 2 bits
	prefixBits := (prefix & 0xC0) >> 6
	remainingBits := prefix & 0x3F
	var result string
	switch prefixBits {
	// 6 bits length
	case 0:

		// masking with 0x3F
		// 0011111 -> binary representation
		// we will get last 6 bits
		len := int(remainingBits)
		buffer, error := r.ReadNBytes(len)
		result, err = string(buffer), error
	// 14 bits length
	case 1:

		lenBuffer := make([]byte, 2)
		lenBuffer[0] = remainingBits

		secondByte, err := r.ReadByte()
		if err != nil {
			return "", err
		}

		lenBuffer[1] = secondByte

		len := int(binary.BigEndian.Uint16(lenBuffer))
		buffer, error := r.ReadNBytes(len)
		result, err = string(buffer), error

	// 32 bits
	case 2:

		lenBuffer, err := r.ReadNBytes(4)
		if err != nil {
			return "", err
		}

		len := binary.BigEndian.Uint32(lenBuffer)
		buffer, error := r.ReadNBytes(int(len))
		result, err = string(buffer), error
	case 3:
		switch remainingBits {
		// 8 bit Integer string
		case 0:
			buffer, err := r.ReadByte()
			if err != nil {
				return "", err
			}
			result, err = fmt.Sprint(buffer), nil
			// 16 Bits Integer String
		case 1:
			buffer, err := r.ReadNBytes(2)
			if err != nil {
				return "", err
			}
			result, err = fmt.Sprint(binary.LittleEndian.Uint16(buffer)), nil
			// 32 bit Integer String
		case 2:
			buffer, err := r.ReadNBytes(4)
			if err != nil {
				return "", err
			}
			result, err = fmt.Sprint(binary.LittleEndian.Uint32(buffer)), nil

		}
	}

	return result, err
}

func (r *RDBReader) ReadNBytes(n int) ([]byte, error) {

	buffer := make([]byte, n)
	_, err := io.ReadFull(r.reader, buffer)
	return buffer, err
}

func (r *RDBReader) ReadByte() (byte, error) {
	var (
		data byte
		err  error
	)

	data, err = r.reader.ReadByte()
	return data, err
}

func (r *RDBReader) Close() {
	if r.file == nil {
		return
	}
	r.file.Close()
}
