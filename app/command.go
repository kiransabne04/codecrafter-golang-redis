package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type CommandFunc func(s *Server, c net.Conn, args []string) string
const emptyHexStr = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
// type Record struct {
// 	Value     any
// 	CreatedAt time.Time
// 	ExpiresAt time.Time
// }

var CommandMap = map[string]func(s *Server, c net.Conn, args []string)string{
	"PING": (*Server).PingCommand,
	"ECHO": (*Server).EchoCommand,
	"SET":  (*Server).SetCommand,
	"GET":  (*Server).GetCommand,
	"INFO": (*Server).InfoCommand,
	"REPLCONF": (*Server).ReplConfCommand,
	//"FULLRESYNC": (*Server).FullResyncCommand,
	"PSYNC": (*Server).PsyncCommand,
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

func (s *Server) PingCommand(c net.Conn, args []string) string {
	if len(args) > 0 {
		return "-ERR wrong number of argument for 'PING' command\r\n"
	}
	
	return fmt.Sprintf("+PONG\r\n")
}

func (s *Server) ReplConfCommand(c net.Conn, args []string) string {
	fmt.Println("args -< ", args)
	if len(args) == 1 {
		return "-ERR wrng number of 'REPLCONF' arguments\r\n"
	}
	return fmt.Sprintf("+OK\r\n")
}

// reponse for PSYNC command received from slave to master
func (s *Server)PsyncCommand(c net.Conn, args[] string) string {
	fmt.Println("full resync command received:", len(args))
	if len(args) > 2 {
		return "-ERR wrong number of 'PSYNC' arguments\r\n"
	}
	//return fmt.Sprintf("+FULLRESYNC %s 0\r\n", s.MasterReplid)
	// reposne with FullRESYNC command & then later with rdb file
	fullResyncResponse := fmt.Sprintf("+FULLRESYNC %s 0\r\n", s.MasterReplid)
	_, err := c.Write([]byte(fullResyncResponse))
	if err != nil {
		fmt.Println("error sending resync response")
		return "-ERR error sending resync response"
	}

	// asending file to server
	var emptyRDB, _ = hex.DecodeString(emptyHexStr)
	_, err = c.Write(append([]byte(fmt.Sprintf("$%d\r\n", len(emptyRDB))), emptyRDB...))
	if err != nil {
		fmt.Println("error sending rdb file")
		return "-ERR error sending rdb file"
	}
	// After sending the RDB, the master should wait for further commands or propagate writes
	fmt.Println("RDB file sent successfully")

	s.ConnectedReplica = c
	s.Replica[c] = true
	
	fmt.Println("s.ConnectedReplica -> ", s.ConnectedReplica.RemoteAddr(), s.ConnectedReplica.LocalAddr(), s.Replica)
	return ""
	//return s.fullSync(c)
}


func (s *Server) EchoCommand(c net.Conn, args []string) string {
	if len(args) != 1 {
		return "-ERR wrng number of arguments for 'ECHO' command\r\n"
	}
	fmt.Println("echo command -> ", args, args[0])
	return fmt.Sprintf("+%s\r\n", args[0])
}

func (s *Server) SetCommand(c net.Conn, args []string) string {
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'SET' command\r\n"
	}
	key := args[0]
	value := args[1]
	record := &Record{
		Value: value,
		CreatedAt: time.Now(),
	}
	// check if PX is used for setting expiary of the key. If set then validate it and convert the millis into int and set it for records values
	if len(args) == 4 && strings.ToUpper(args[2]) == "PX" {
		expiration, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Println(err)
			return "-ERR wrong expiration time provided for the record"
		}
		record.ExpiresAt = time.Now().Add(time.Duration(expiration) * time.Millisecond)
	}

	s.DataStore[key] = record
	fmt.Println("set commabd -> ", s.ConnectedReplica)
	if s.ConnectedReplica != nil {
		fmt.Println("replica set command inside set method")
		s.propagateCommandToReplica("SET", args)
	}
	return fmt.Sprintf("+OK\r\n")
}

func (s *Server) GetCommand(c net.Conn, args []string) string {
	if len(args) != 1 {
		return "-ERR wrong number of arguments for 'GET' command\r\n"
	}
	key := args[0]
	val, prsnt := s.DataStore[key]
	if !prsnt {
		return "$-1\r\n"
	}

	// check for expiary
	if time.Now().After(val.ExpiresAt) && !val.ExpiresAt.IsZero() {
		delete(s.DataStore, key)
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(val.Value.(string)), val.Value)
}

func (s *Server) InfoCommand(c net.Conn, args []string) string {
	if len(args) != 1 && strings.ToLower(args[1]) != "replication" {
		return "-ERR wrong number of arguments for 'INFO' command"
	}
	replicationInfo := fmt.Sprintf(`# Replication
	role:%s
	connected_slaves:%d
	master_replid:%s
	master_repl_offset:%d
	second_repl_offset:-1
	repl_backlog_active:%d
	repl_backlog_size:%d
	repl_backlog_first_byte_offset:%d
	repl_backlog_histlen:%d
	`, s.Role, s.ConnectedSlaves, s.MasterReplid, s.MasterReplOffset, s.ReplBacklogActive, s.ReplBacklogSize, s.ReplBacklogFirstByteOffset, s.ReplBacklogHistlen )

	return fmt.Sprintf("$%d\r\n%s\r\n", len(replicationInfo), replicationInfo)
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

func (s *Server) propagateCommandToReplica(command string, args []string) {
	cmd := fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", len(args) + 1, len(command), command)
	for _, arg := range args {
		cmd += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}
	fmt.Println("cmd -> ", cmd, args)
	//send the comand to replica
	// _, err := s.ConnectedReplica.Write([]byte(cmd))
	// if err != nil {
	// 	fmt.Println("Error sending command to replica")
	// 	return
	// }

	//propagate commands to all connected replicas
	for replica := range s.Replica{
		fmt.Println("replica ->", replica, cmd)

		_, err := replica.Write([]byte(cmd))
		if err != nil {
			fmt.Println("error sending command tio replica")
			delete(s.Replica, replica) //remove replica if connection is not their
		}
	}
	fmt.Printf("Propagated command to replica %s %v\n", command, args)
}

