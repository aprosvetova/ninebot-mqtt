package protocol

type RequestCommand byte

type Response struct {
	Command byte
	Parameter byte
	Payload []byte
}
