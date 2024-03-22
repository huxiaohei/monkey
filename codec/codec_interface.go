package codec

type Codec interface {
	CodecName() string
	Encode(message interface{}) ([]byte, error)
	Decode(data []byte) (length int, message interface{}, err error)
}
