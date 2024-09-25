package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

// inputs for command cli flags
var (
	port = flag.String("port", "6379", "port number to connect on")
	addr = flag.String("addr", "0.0.0.0", "addressr to connect on") 
	replicaof = flag.String("replicaof", "", "replica flag for the cli")
)

// server struct containing server properties & stats
type Server struct {
	Role string
	Listener net.Listener
	Stats ServerStats
	DataStore map[string]*Record
	ConnectedSlaves            int
	MasterReplid               string
	MasterReplOffset           int64
	ReplBacklogActive          int
	ReplBacklogSize            int
	ReplBacklogFirstByteOffset int64
	ReplBacklogHistlen         int64
	MasterHost string
	MasterPort string
	ConnectedReplica	net.Conn
}

// serverstats struct contains the stats of the server, like connectioncounts, commands processed etc.
type ServerStats struct {
	ConnectionCount int
	CommandsProcessed int
}

// record dt
type Record struct {
	Value any
	CreatedAt time.Time
	ExpiresAt time.Time
}

// initializes the new server with default config or opts
func NewServer() *Server {
	
	server := &Server{}
	
	// if replicaof is provided then take masterHost & masterPort

	server.Stats = ServerStats{}
	server.DataStore = make(map[string]*Record)
	// server.MasterReplid = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	// server.MasterReplOffset = 0
	return server
}

// start server function starts the server and listens to tge mentioned port
func (s *Server) startServer (host, port, replicaof string) error {
	var err error
	fmt.Println("Server is starting to initialize")

	s.Listener, err = net.Listen("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		fmt.Sprintf("failed to bind port %s \n", addr)
		return err
	}
	fmt.Sprintf("s.Listener %s\n", addr)

	// if replicaof is empty, we are starting as master else slave
	if replicaof == "" {
		s.Role = "master"
		s.MasterHost = host
		s.MasterPort = port
		s.MasterReplid = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		s.MasterReplOffset = 0
		fmt.Printf("Server is running as master on %s:%s\n", host, port)
		//s.HandShakeCommand()
	} else {
		var masterHost, masterPort string
		fmt.Sscanf(replicaof, "%s %s", &masterHost, &masterPort)
		s.Role = "slave"
		s.MasterHost = masterHost
		s.MasterPort = masterPort
		fmt.Printf("Server running as slave connecting to master at %s:%s\n", masterHost, masterPort)
		//s.HandShakeCommand()
	}
	

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Println("error accepting connections ", err)
			continue
		}
		go s.handleConn(conn)
		// go func (c net.Conn)  {
		// 	reader := bufio.NewReader(c)
		// 	line, _ := reader.Peek(4) // peek at first few bytes of the connection

		// 	// If the connection is from a replica (expecting PSYNC), start the handshake
		// 	if string(line) == "PSYN" {
		// 		fmt.Println("Replica connection detected, starting handshake.")
		// 		s.HandShakeCommand(c)
		// 	} else {
		// 		// Otherwise, assume it's a client connection
		// 		fmt.Println("Client connection detected, processing commands.")
		// 		s.handleConn(c)
		// 	}
		// }(conn)
		
	}

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	flag.Parse()
	server := NewServer()

	err := server.startServer(*addr, *port, *replicaof)
	if err != nil {
		fmt.Printf("Failed to start the server: %v\n", err)
		os.Exit(1)
	}

}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	line, _ := reader.Peek(4)
	fmt.Println("line ->:",string(line))
	// reader receives inputs & is parsed with parseCommand function. This also converts the incoming resp protocal commands to slice of string & returns error if any
	for {
		inputCmd, err := parseCommand(reader)
		
		if err != nil {
			if err == io.EOF{
				fmt.Println("connection is closed")
			} else {
				fmt.Println("Invalid command received")
				conn.Write([]byte("-ERR invalid commands \r\n"))
			}
			break
		}
		fmt.Println("inputCmd -> ", inputCmd, err)
		// csacascas
		fmt.Println(inputCmd[0] == "PSYNC")
		// if inputCmd[0] == "PSYNC" {
		// 	s.HandShakeCommand(conn)
		// }
		response  := executeCommand(s, conn, inputCmd)
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("error in writing response -. ", err)
			break
		}
	}

}
