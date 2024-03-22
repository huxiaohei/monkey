package codec

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"monkey/codec"
	"monkey/logger"
	"reflect"
)

var (
	_       codec.Codec = &GatewayCodec{}
	mlog, _             = logger.GetLoggerManager().GetLogger(logger.MainTag)
)

type GatewayCodec struct {
}

func (ec *GatewayCodec) CodecName() string {
	return "GatewayCodec"
}

func (ec *GatewayCodec) Encode(message interface{}) (buff []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			mlog.Errorf("GatewayCodec encode panic: %v", e)
			err = fmt.Errorf("GatewayCodec encode panic: %v", e)
			buff = nil
		}
	}()

	prefix := []byte(ec.CodecName())
	var buffer bytes.Buffer
	err = binary.Write(&buffer, binary.BigEndian, int32(len(prefix)))
	if err != nil {
		return nil, err
	}
	buffer.Write(prefix)

	msgName := reflect.TypeOf(message).Name()
	err = binary.Write(&buffer, binary.BigEndian, int32(len(msgName)))
	if err != nil {
		return nil, err
	}
	buffer.Write([]byte(msgName))

	encoder := json.NewEncoder(&buffer)
	err = encoder.Encode(message)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (ec *GatewayCodec) Decode(data []byte) (lenght int, message interface{}, err error) {
	return len(data), string(data), nil
}
