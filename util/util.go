// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"io"
	"regexp"
)

const (
	sevenBitMask   uint32 = 0x7f
	mostSigBit     uint32 = 0x80
	tableSeparator rune   = '/'
)

//EncodeULeb128 encodes the integer using the LEB128 Standard.
//32 bit unsigned integer should be sufficient as that would represent a data length of 4.2 GB.
func EncodeULeb128(value uint32, writer io.Writer) error {
	//figure out the remaining bits
	remaining := value >> 7
	//If there are still bits available
	for remaining != 0 {
		//Grabs the 7 bits from value then prepends a 1
		_, err := writer.Write([]byte{byte(value&sevenBitMask | mostSigBit)})
		if err != nil {
			return err
		}
		//Adjust the new value
		value = remaining
		//Adjust the new remaining
		remaining >>= 7
	}
	//Write the last bit of the int without prepended 0
	_, err := writer.Write([]byte{byte(value & sevenBitMask)})
	return err
}

//DecodeULeb128 decodes from the reader using the LEB128 standard and returns a uint32 of the value it found.
func DecodeULeb128(reader io.Reader) (uint32, error) {
	var result uint32
	var cur = [1]byte{0x80}
	var ctr uint32
	//While there is always more to read
	for cur[0]&byte(mostSigBit) == byte(mostSigBit) {
		//Read the value.
		//If this is the last value (mostSigBit is 0) then this will break loop in next iteration.
		_, err := io.ReadFull(reader, cur[:])
		if err != nil {
			return 0, err
		}
		result += uint32(cur[0]&byte(sevenBitMask)) << (ctr * 7)
		ctr++
	}
	return result, nil
}

//SanitizeKey with make sure a "/" exist before the key but not after.
func SanitizeKey(key string) string {
	// Keys with multiple slashes into a single slash (eg. '//', '///', etc to '/')
	re := regexp.MustCompile(`/+`)
	key = re.ReplaceAllString(key, "/")

	// convert string to list of runes (characters)
	sanitized := []rune(key)

	// ensure first character is '/'
	if sanitized[0] != tableSeparator {
		sanitized = append([]rune{tableSeparator}, sanitized...)
	}

	// ensure last character is not '/'
	if sanitized[len(sanitized)-1] == tableSeparator {
		sanitized = sanitized[:len(sanitized)-1]
	}

	// convert character array back to string
	return string(sanitized)
}
