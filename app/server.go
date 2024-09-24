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
	if *replicaof == "" {
		server.Role = "master"
	} else {
		server.Role = "slave"
	}
	// if replicaof is provided then take masterHost & masterPort
	var masterHost, masterPort string
	fmt.Sscanf(*replicaof, "%s %s", &masterHost, &masterPort)

	server.Stats = ServerStats{}
	server.DataStore = make(map[string]*Record)
	server.MasterReplid = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	server.MasterReplOffset = 0
	server.MasterHost = masterHost
	server.MasterPort = masterPort

	fmt.Println("masterHost & port -> ", masterHost, masterPort)
	return server
}

// start server function starts the server and listens to tge mentioned port
func (s *Server) startServer (addr string) error {
	var err error
	fmt.Println("Server is starting to initialize")

	s.Listener, err = net.Listen("tcp", addr)
	if err != nil {
		fmt.Sprintf("failed to bind port %s \n", addr)
		return err
	}
	if s.Role == "slave" {
		s.HandShakeCommand()
	}
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Println("error accepting connections ", err)
			continue
		}
		go s.handleConn(conn)
	}

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	flag.Parse()
	// l, err := net.Listen("tcp", "0.0.0.0:6379")
	// if err != nil {
	// 	fmt.Println("Failed to bind port 6379")
	// 	os.Exit(1)
	// }

	// for {
	// 	conn, err := l.Accept()
	// 	if err != nil {
	// 		fmt.Println("Error accepting connection: ", err.Error())
	// 		continue
	// 	}
		
	// 	go handleConn(conn)
	// }
	server := NewServer()

	if err := server.startServer(fmt.Sprintf("%s:%s", *addr, *port)); err != nil {
		fmt.Println("error starting the server ", err)
		os.Exit(1)
	}

}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
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
		response  := executeCommand(s, conn, inputCmd)
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("error in writing response -. ", err)
			break
		}

		
	}
}
