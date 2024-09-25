package main

import (
	"fmt"
	"net"
	"strings"
)


func (s *Server)HandShakeCommand() {
	fmt.Println(" HandShakeCommand called")
	
	address := fmt.Sprintf("%s:%s", s.MasterHost, s.MasterPort)
	m, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("-ERR couldnt connect to master at "+ address )
	}
	defer m.Close()
	// sned PING to the master
	_, err = m.Write([]byte("*1\r\n$4\r\nPING\r\n"))
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
	// //check if response is PONG
	if strings.TrimSpace(string(response)) != "PONG" {
		fmt.Println("-ERR invalid response from master")
		return
	}

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

	// //time.Sleep(1 * time.Second)
	// m.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$4\r\n6380\r\n"))
	// //return fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", 1, 4, "PING")
	// //time.Sleep(1 * time.Second)
	// m.Write([]byte("*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"))

}

