package message

//ClientHelloComplete is sent to the server after the client finished giving the server all of its entries that does not already exist on the server.
type ClientHelloComplete struct {
	message
}

//NewClientHelloComplete creates a new instance of ClientHelloComplete
func NewClientHelloComplete() *ClientHelloComplete {
	return &ClientHelloComplete{
		message{
			mType: mTypeClientHelloComplete,
		},
	}
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (chc *ClientHelloComplete) MarshalMessage() ([]byte, error) {
	var output []byte
	return append(output, chc.Type()), nil
}