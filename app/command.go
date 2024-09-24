package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type CommandFunc func(s *Server, c net.Conn, args []string) string

// type Record struct {
// 	Value     any
// 	CreatedAt time.Time
// 	ExpiresAt time.Time
// }

var CommandMap = map[string]CommandFunc{
	"PING": PingCommand,
	"ECHO": EchoCommand,
	// "SET":  SetCommand,
	// "GET":  GetCommand,
	// "INFO": InfoCommand,
}

// *1\r\n$4\r\nPING\r\n
// *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n -> ["ECHO", "hey"]
// parse command which parses the incoming inputs bytes from cli and connections. It takes bufio reader as inputs and returns slice of string & error
func parseCommand(reader *bufio.Reader) ([]string, error) {
	// divide the incoming reader bytes and get line for each line break
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error in reading from reader")
		return nil, err
	}
	// trimspaces of the line
	line = strings.TrimSpace(line)
	
	// check if its only an empty string
	if line == "" {
		fmt.Println("recevied empty command ")
		return nil, errors.New("-ERR received empty commands")
	}
	fmt.Println("line -> ", line, string(line[0]))
	// now check the first character of the line and then apply parsing accordingly based on resp protocol
	switch line[0] {
	case '+', '-', ':', '$':
		return []string{line}, nil //return it as it is
	case '*':
		return parseArray(reader, line) // parse the array

	default:
		return nil, errors.New("invalid command Received")

	}
}

// this function parses the input into array of string commands based on the reader and line. considering the example command *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n -> ["ECHO", "hey"]
// in parse command func we checked that line[0] is '*' character, so it indicates that 2 array inputs is received as commands. so we will parse the number of args first, then continue accordingly
func parseArray(reader *bufio.Reader, line string) ([]string, error){
	fmt.Println("line -> ", strings.TrimSpace(line[1:]))
	numArgs, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("invalid num of args received")
	}
	fmt.Println("numArgs -> ", numArgs)
	args := make([]string, numArgs)
	for i := 0; i < numArgs; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		// check if line is empty string
		if line == "" {
			return nil, errors.New("-ERR recevied empty args")
		}
		//check first character of line[0] again, if its $ then the next character will be number denoting lenght of Bulk string
		switch line[0] {
		case '$':
			length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
			if err != nil {
				return nil, err
			}
			//make buf input of the required length got above and readfull from the reader into it
			inputBuf := make([]byte, length)
			_, err = io.ReadFull(reader, inputBuf)
			if err != nil {
				return nil, err
			}

			//read trailing \n
			_, err = reader.ReadString('\n')
			if err != nil {
				return nil, err
			}

			args[i] = string(inputBuf)

		default:
			return nil, errors.New("invalid bulk string format")
		}

	}
	return args, nil
}

func PingCommand(s *Server, c net.Conn, args []string) string {
	if len(args) > 0 {
		return "-ERR wrong number of argument for 'PING' command\r\n"
	}
	return fmt.Sprintf("+PONG\r\n")
}

func EchoCommand(s *Server, c net.Conn, args []string) string {
	if len(args) != 1 {
		return "-ERR wrng number of arguments for 'ECHO' command\r\n"
	}
	fmt.Println("echo command -> ", args, args[0])
	return fmt.Sprintf("+%s\r\n", args[0])
}

func executeCommand(s *Server, c net.Conn, args []string) string {
	fmt.Println("execute commands -> ", args, args[0], args[0][0])
	if len(args) == 0 {
		return "-ERR no command provided"
	}

	fmt.Println("execute command received -> ", args)
	switch args[0][0] {
	case '+':
		return fmt.Sprintf("%s\r\n", args[0])

	case '-':
		return fmt.Sprintf("%s\r\n", args[0])

	case ':':
		return fmt.Sprintf("%s\r\n", args[0])

	case '$':
		
		length, err := strconv.Atoi(strings.TrimSpace(args[0][1:]))
		if err != nil {
			return "-ERR invalid bulk string length"
		}
		fmt.Println("execute command $ -> ", args[0][0], args[0][1], strings.TrimSpace(args[0][1:]), length)
		return fmt.Sprintf("%d\r\n%s\r\n", length, args[1])

	default:
		commands := strings.ToUpper(args[0])
		fmt.Println("e xommands -< ", args[0], commands)
		if commandFunc, exists := CommandMap[commands]; exists {
			return commandFunc(s, c, args[1:])
		}
		return "-ERR Unknown Command"
	}
}



