package message

import (
	"bytes"
	"io"
	"testing"
)

func TestClearAllEntries_UnmarshalMessage(t *testing.T) {
	type fields struct {
		message message
		magic   [4]byte
		valid   bool
	}
	type args struct {
		reader io.Reader
	}
	expectedMagic := ClearAllMagic()
	tests := []struct {
		name       string
		args       args
		wantFields fields
		wantErr    bool
	}{
		{
			name: "valid message",
			args: args{bytes.NewBuffer(expectedMagic[:])},
			wantFields: fields{
				message: message{
					mType: MTypeClearAllEntries,
				},
				magic: expectedMagic,
				valid: true,
			},
			wantErr: false,
		},
		{
			name: "invalid message",
			args: args{bytes.NewBuffer([]byte{0x00, 0x11, 0x22, 0x33})},
			wantFields: fields{
				message: message{
					mType: MTypeClearAllEntries,
				},
				magic: [4]byte{0x00, 0x11, 0x22, 0x33},
				valid: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cae := &ClearAllEntries{}
			if err := cae.UnmarshalMessage(tt.args.reader); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
			if cae.mType != tt.wantFields.message.mType {
				t.Errorf("mType: want = %v, got = %v", tt.wantFields.message.mType, cae.mType)
			}
			if cae.magic != tt.wantFields.magic {
				t.Errorf("magic: want = %v, got = %v", tt.wantFields.magic, cae.magic)
			}
			if cae.valid != tt.wantFields.valid {
				t.Errorf("valid: want = %v, got = %v", tt.wantFields.valid, cae.valid)
			}
		})
	}
}
