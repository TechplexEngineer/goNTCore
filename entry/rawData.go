package entry

import (
	"io"

	"github.com/technomancers/goNTCore/util"
)

//RawData is any type that is not covered in the protocole.
//It is recommended to add a byte or two to the beginning to describe what type it is.
type RawData struct {
	entry
	data []byte
}

//NewRawData creates and instance of RawData.
func NewRawData(data []byte) *RawData {
	return &RawData{
		entry: entry{
			eType: eTypeRawData,
		},
		data: data,
	}
}

//MarshalEntry implements Marshaler for Network Table Entry.
func (rd *RawData) MarshalEntry(writer io.Writer) error {
	err := util.EncodeULeb128(uint32(len(rd.data)), writer)
	if err != nil {
		return err
	}
	_, err = writer.Write(rd.data)
	return err
}

//UnmarshalEntry implements Unmarshaler for Network Table Entry.
func (rd *RawData) UnmarshalEntry(reader io.Reader) error {
	rd.eType = eTypeRawData
	dataLength, err := util.DecodeULeb128(reader)
	if err != nil {
		return err
	}
	buf := make([]byte, dataLength)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return err
	}
	rd.data = buf
	return nil
}
