package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
    STRING  = '+'
    ERROR   = '-'
    INTEGER = ':'
    BULK    = '$'
    ARRAY   = '*'

    TYPE_BULK = "Bulk"
    TYPE_ARRAY = "Array"
)

type Value struct {
    typ   string  // data type of the value
    str   string  // holds value of string from simple strings
    num   int     // holds value of int from integers
    bulk  string  // holds value of strings from bulk strings
    array []Value // holds all valyes recieved from arrays
}


type Resp struct {
    reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
    return &Resp{reader: bufio.NewReader(rd)}
}

// reads one byte at a time until we reach '\r' indicating the end of a line 
// then, we return the line w/o the last two bytes \r\n, and n number of bytes in the line
func (r *Resp) readLine() (line []byte, n int, err error) {
    for {
        // read the byte
        b, err := r.reader.ReadByte()
        if err != nil {
            return nil, 0, err
        }

        // inc count and add byte to []byte
        n += 1
        line = append(line, b)
        if len(line) >= 2 && line[len(line)-2] == '\r' {
            break
        }
    }
    return line[:len(line)-2], n, nil
}

// parse single integer from the reader
// x (integer) n (num bytes read) err (error)
func (r *Resp) readInteger() (x int, n int, err error) {

	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()

	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// Skip the first byte because we have already read it in the Read method.
// Read the integer that represents the number of elements in the array.
// Iterate over the array and for each line, call the Read method to parse the type according to the character at the beginning of the line.
// With each iteration, append the parsed value to the array in the Value object and return it

// method to recursively read array
func (r *Resp) readArray() (Value, error) {

    // skip the first byte
    r.reader.ReadByte()

    // Read the integer representing n of elements in the array
    length, _, err := r.readInteger()
    if err != nil {
        panic("readArray: Could not read n from []byte")
    }


    // iterate over the array and call read on each element and append it to the val
    var val Value

    for _ = range length {
        readVal, err := r.Read()
        if err != nil {
            return val, err
        }

        val.array = append(val.array, readVal) 
    }

   return val, nil 
}

// Skip the first byte because we have already read it in the Read method.
// Read the integer that represents the number of bytes in the bulk string.
// Read the bulk string, followed by the ‘\r\n’ that indicates the end of the bulk string.
// Return the Value object.

// method to recursively read Bulk ?
func (r *Resp) readBulk() (Value, error){
    v := Value{}
    v.typ = "bulk"
    // skip the first byte since we've already read it
    r.reader.ReadByte()

    
    // read the integer that represents the number of bytes in the bulk string
    length, _, err := r.readInteger()
    if err != nil {
        return v, err
    }
    
    // allocate []byte for bulk
    bulk := make([]byte, length)
    
    // read from reader into bulk "records data into p"
    r.reader.Read(bulk)

    v.bulk = string(bulk)

    // read off \r\n
    r.readLine()

    return v, nil
}
