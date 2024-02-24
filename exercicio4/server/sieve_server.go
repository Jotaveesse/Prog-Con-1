package main

import (
	"encoding/json"
	"exercicio4/server/service"
	"exercicio4/shared"
	"fmt"
	"math"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

func main() {
	var conn_type string

	for conn_type != "u" && conn_type != "t" && conn_type != "r" {
		fmt.Print("Choose (u) -> udp | (t) -> tcp | (r) -> rpc:  ")
		fmt.Scan(&conn_type)
	}

	if conn_type == "u" {
		SieveServerUDP()
	} else if conn_type == "t" {
		SieveServerTCP()
	} else {
		SieveServerRPC()
	}

	fmt.Scanln()
}

func SieveServerRPC() {
	// cria uma instância da calculadora de crivo
	sieveCalculator := new(service.SieveCalcRPC)

	// cria um novo servidor RPC e registra a calculadora de crivo
	server := rpc.NewServer()
	err := server.Register(sieveCalculator)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// cria um listener TCP
	ln, err := net.Listen("tcp", ":1313")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	defer ln.Close()

	// aguarda por invocações
	fmt.Println("Servidor RPC aguardando invocações na porta 1313...")
	server.Accept(ln)
}

func SieveServerTCP() {

	//  define o endpoint do servidor TCP
	r, err := net.ResolveTCPAddr("tcp", "localhost:1314")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// cria um listener TCP
	ln, err := net.ListenTCP("tcp", r)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	fmt.Println("Servidor TCP aguardando conexões na porta 1314...")

	for {
		// aguarda/aceita conexão
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		// processa requests da conexão
		go HandleTCPConnection(conn)
	}
}

func HandleTCPConnection(conn net.Conn) {
	var msgFromClient shared.Request

	// Close connection
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			if opErr, ok := err.(*net.OpError); !ok || opErr.Err != net.ErrClosed {
				fmt.Println(err)
				os.Exit(0)
			}
		}
	}(conn)

	// Cria coder/decoder JSON
	jsonDecoder := json.NewDecoder(conn)
	jsonEncoder := json.NewEncoder(conn)

	for {
		// recebe & unmarshall requests do cliente
		err := jsonDecoder.Decode(&msgFromClient)
		if err != nil && err.Error() == "EOF" {
			conn.Close()
			break
		}

		// processa request
		r := service.SieveCalc{}.InvokeSieveCalc(msgFromClient)

		// cria resposta
		msgToClient := shared.Reply{Result: r}

		// serializa & envia resposta para o cliente
		err = jsonEncoder.Encode(msgToClient)
		if err != nil {
			fmt.Println(err)
			break
			//os.Exit(0)

		}
		//fmt.Println("Sent response with", len(r), "primes")
	}
}

func SieveServerUDP() {
	msgFromClient := make([]byte, 1024)

	// resolve server address
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(shared.SievePort))
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// listen on udp port
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// close conn
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}(conn)

	fmt.Println("Server UDP is ready to accept requests at port", shared.SievePort, "...")

	for {
		// receive request
		n, addr, err := conn.ReadFromUDP(msgFromClient)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		// handle request
		HandleUDPRequest(conn, msgFromClient, n, addr)
	}
}

func HandleUDPRequest(conn *net.UDPConn, msgFromClient []byte, n int, addr *net.UDPAddr) {
	var msgToClient []byte
	var request shared.Request

	//unmarshall request
	err := json.Unmarshal(msgFromClient[:n], &request)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// process request
	r := service.SieveCalc{}.InvokeSieveCalc(request)

	// create response
	rep := shared.Reply{Result: r}

	// serialise response
	msgToClient, err = json.Marshal(rep)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	var ChunkSize = 1024

	packetCount := math.Ceil(float64(len(msgToClient)) / float64(ChunkSize))
	packetsToClient, err := json.Marshal(packetCount)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	_, err = conn.WriteTo(packetsToClient, addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	start := time.Now()

	// divide mensagem em varrios pendaços e envia cada um deles
	for i := 0; i < len(msgToClient); i += ChunkSize {
		end := i + ChunkSize
		if end > len(msgToClient) {
			end = len(msgToClient)
		}
		chunk := msgToClient[i:end]

		// send a chunk
		_, err := conn.WriteTo(chunk, addr)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

		if ((i / 1024) % 60) == 59 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	end := time.Now()
	fmt.Println(end.Sub(start))

	//fmt.Println("Sent response with", len(r), "primes")
}
