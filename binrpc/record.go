package binrpc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Record represents a BINRPC type+size, and Go value. It is not a binary representation of a record.
// Type is the BINRPC type.
type Record struct {
	size int

	Type  uint8
	Value any
}

// StructItem represents an item in a BINRPC struct. Because BINRPC structs may contain the same key multiple times,
// structs are handled with arrays of StructItem.
type StructItem struct {
	Key   string
	Value Record
}

// String returns the string value, or an error if the type is not a string.
func (record Record) String() (string, error) {
	if record.Type != TypeString {
		return "", fmt.Errorf("type error: expected type string (%d), got %d", TypeString, record.Type)
	}

	return record.Value.(string), nil
}

// Int returns the int value, or an error if the type is not a int.
func (record Record) Int() (int, error) {
	if record.Type != TypeInt {
		return 0, fmt.Errorf("type error: expected type int (%d), got %d", TypeInt, record.Type)
	}

	return record.Value.(int), nil
}

// Double returns the double value as a float64, or an error if the type is not a double
func (record Record) Double() (float64, error) {
	if record.Type != TypeDouble {
		return 0, fmt.Errorf("type error: expected type double (%d), got %d", TypeDouble, record.Type)
	}

	return record.Value.(float64), nil
}

// StructItems returns items for a struct value, or an error if not a struct.
func (record *Record) StructItems() ([]StructItem, error) {
	if record.Type != TypeStruct {
		return nil, fmt.Errorf("type error: expected type struct (%d), got %d", TypeStruct, record.Type)
	}

	return record.Value.([]StructItem), nil
}

// Scan copies the value in the Record into the values pointed at by dest. Valid dest type are *int, *string, and *[]StructItem
func (record *Record) Scan(dest any) error {
	switch dest.(type) {
	case *string:
		s := dest.(*string)

		switch record.Type {
		case TypeString:
			*s = record.Value.(string)
		case TypeInt:
			*s = strconv.Itoa(record.Value.(int))
		case TypeDouble:
			*s = fmt.Sprintf("%.3f", record.Value.(float64))
		default:
			return fmt.Errorf("type error: cannot convert type %d to string", record.Type)
		}
	case *int:
		i := dest.(*int)

		switch record.Type {
		case TypeString:
			if n, err := strconv.Atoi(record.Value.(string)); err == nil {
				*i = n
			} else {
				return err
			}
		case TypeInt:
			*i = record.Value.(int)
		default:
			return fmt.Errorf("type error: cannot convert type %d to int", record.Type)
		}
	case *float64:
		f := dest.(*float64)

		switch record.Type {
		case TypeString:
			if value, err := strconv.ParseFloat(record.Value.(string), 64); err == nil {
				*f = value
			} else {
				return err
			}
		case TypeInt:
			*f = float64(record.Value.(int))
		case TypeDouble:
			*f = record.Value.(float64)
		default:
			return fmt.Errorf("type error: cannot convert type %d to double", record.Type)
		}
	case *[]StructItem:
		if record.Type != TypeStruct {
			return fmt.Errorf("type error: cannot convert type %d to []StructItem", record.Type)
		}

		items := dest.(*[]StructItem)
		*items = record.Value.([]StructItem)
	default:
		return errors.New("invalid dest type")
	}

	return nil
}

// Encode is a low level function that encodes a record and writes it to w.
func (record *Record) Encode(w io.Writer) error {
	var value bytes.Buffer

	switch record.Type {
	case TypeInt:
		var v int
		var ok bool

		if v, ok = record.Value.(int); !ok {
			return errors.New("type error: expected type int")
		}

		// shortcut!
		if v == 0 {
			_, err := w.Write([]byte{0x00})
			return err
		}

		value.Write(intToBytesBE(v))
	case TypeString:
		if s, ok := record.Value.(string); !ok {
			return errors.New("type error: expected type string")
		} else {
			value.WriteString(s)
		}

		value.WriteByte(0x00)
	case TypeDouble:
		var v float64
		var ok bool

		if v, ok = record.Value.(float64); !ok {
			return errors.New("type error: expected type float64")
		}

		value.Write(intToBytesBE(int(v * 1000)))
	default:
		return fmt.Errorf("type error: type %d not implemented", record.Type)
	}

	sizeOfValue := value.Len()

	var buffer bytes.Buffer

	if sizeOfValue < 8 {
		// this can fit in 3 bits
		header := byte(sizeOfValue<<4) | record.Type
		buffer.WriteByte(header)
		buffer.Write(value.Bytes())
	} else {
		sizeBytes := intToBytesBE(sizeOfValue)

		header := 1<<7 | uint8(len(sizeBytes)<<4) | record.Type

		buffer.WriteByte(header)
		buffer.Write(sizeBytes)
		buffer.Write(value.Bytes())
	}

	if _, err := buffer.WriteTo(w); err != nil {
		return err
	}

	return nil
}
