package main

import (
	"fmt"
	"net"
	"strings"
)

func (s *Server)HandShakeCommand(c net.Conn) {
	fmt.Println(" HandShakeCommand called")

	// address := fmt.Sprintf("%s:%s", s.MasterHost, s.MasterPort)
	// m, err := net.Dial("tcp", address) // connecting to the master
	// if err != nil {
	// 	fmt.Println("-ERR couldnt connect to master at "+ address )
	// }
	m := c
	//s.ConnectedReplica = m
	fmt.Println("s.ConnectedReplica -> ", m.RemoteAddr())
	//defer m.Close()
	// sned PING to the master

	_, err := m.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		fmt.Println("-ERR sending ping to master server failed")
		return
	}

	// read response from the master for ping command
	response := make([]byte, 2048)
	_, err = m.Read(response)
	if err != nil {
		fmt.Println("-ERR reading PONG response from master failed", err)
		return
	}
	fmt.Println("received from master ->",strings.TrimSpace(string(response)))
	fmt.Println(strings.TrimSpace(string(response)))

	//send Replconf listening port to master
	_, err = m.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	if err != nil {
		fmt.Println("-ERR sneding REPLCONF (listening-port) failed")
		return
	}

	// read replconf response
	_, err = m.Read(response)
	if err != nil {
		fmt.Println("-ERR readng REPLCONF response from master failed", err)
		return
	}
	fmt.Println("received REPLCONF (listening-port) response ", string(response))

	// send REPLCONF capa psync2 to master
	_, err = m.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))
	if err != nil {
		fmt.Println("-ERR sendging REPLCONF (capa psync2) failed")
		return
	}

	// read REPLCONF capa psync2 response
	_, err = m.Read(response)
	if err != nil {
		fmt.Println("-ERR reading REPLCONF (capa psync2) response ", err)
		return
	}
	fmt.Println("Received REPLCONF (capa psync2) response ", string(response))
	// send PSYNC ? -1 command to master
	_, err = m.Write([]byte("*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"))
	if err != nil {
		fmt.Println("-ERR sending PSYNC failed :", err)
		return
	}

	// read PSYNC ? -1 response
	_, err = m.Read(response)
	if err != nil {
		fmt.Println("-ERR reading PSYNC response :", err)
		return
	}

	fmt.Println("recevied PSYNC response :", string(response))

}

// func (s *Server) HandShakeCommand(conn net.Conn) {
// 	reader := bufio.NewReader(conn)

// 	for {
// 		// Read the command from the replica
// 		line, err := reader.ReadString('\n')
// 		if err != nil {
// 			fmt.Println("Error reading from replica:", err)
// 			return
// 		}

// 		fmt.Println("Received from replica:", strings.TrimSpace(line))

// 		// Handle RESP commands from the replica
// 		switch {
// 		case strings.HasPrefix(line, "PING"):
// 			conn.Write([]byte("+PONG\r\n"))

// 		case strings.HasPrefix(line, "REPLCONF"):
// 			conn.Write([]byte("+OK\r\n"))

// 		case strings.HasPrefix(line, "PSYNC"):
// 			replicationID := "replid-12345"
// 			conn.Write([]byte(fmt.Sprintf("+FULLRESYNC %s 0\r\n", replicationID)))
// 			s.sendRDB(conn)

// 			// Set the connection for future command propagation
// 			s.ConnectedReplica = conn
// 			return // Exit the handshake and continue for command propagation
// 		}
// 	}
// }

// func (s *Server) sendRDB(conn net.Conn) {
// 	// Send a mock RDB file to the replica
// 	rdbData := "MOCK_RDB_DATA" // This is just a mock string. Replace it with actual RDB file data.
// 	_, err := conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(rdbData), rdbData)))
// 	if err != nil {
// 		fmt.Println("-ERR sending RDB to replica failed:", err)
// 		return
// 	}
// 	fmt.Println("Sent RDB file to replica.")
// }




