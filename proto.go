package main

import (
	"bytes"
	"fmt"
)

const (
	CommandSet    = "set"
	CommandGet    = "get"
	CommandHello  = "hello"
	CommandClient = "client"
)

type Command interface{}

type SetCommand struct {
	key, val []byte
}

type ClientCommand struct {
	value string
}

type HelloCommand struct {
	value string
}

type GetCommand struct {
	key []byte
}

func respWriteMap(m map[string]string) []byte {
	buf := bytes.Buffer{}

	buf.WriteString("%" + fmt.Sprintf("%d\r\n", len(m)))

	for key, value := range m {
		buf.WriteString(fmt.Sprintf("+%s\r\n", key))
		buf.WriteString(fmt.Sprintf("+%s\r\n", value))
	}

	return buf.Bytes()
}
