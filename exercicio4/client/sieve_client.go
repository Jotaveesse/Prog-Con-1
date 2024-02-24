package client

import (
	"encoding/json"
	"exercicio4/shared"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

func Run() {

	var rng int
	var conn_type, tryAgain, calcType string

	calcType = "blk_conc"

	for conn_type != "u" && conn_type != "t" && conn_type != "r" {
		fmt.Print("Choose (u) -> udp | (t) -> tcp | (r) -> rpc: ")
		fmt.Scan(&conn_type)
	}

	var primes []int
	var rtt time.Duration

	if conn_type == "u" {
		conn := StartConnectionUDP()

	connLoopUDP:
		for {
			tryAgain = ""

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt = SendMessageUDP(conn, rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)

			fmt.Print("Want to try again (y) -> yes | (n) -> no: ")
			fmt.Scan(&tryAgain)

			if tryAgain == "n" {
				CloseConnectionUDP(conn)
				break connLoopUDP
			}
		}
	} else if conn_type == "t" {
		conn := StartConnectionTCP()

	connLoopTCP:
		for {
			tryAgain = ""
			
			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt = SendMessageTCP(conn, rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)

			fmt.Print("Want to try again (y) -> yes | (n) -> no: ")
			fmt.Scan(&tryAgain)

			if tryAgain == "n" {
				CloseConnectionTCP(conn)
				break connLoopTCP
			}
		}
	} else {
	coonLoopRPC:
		for {
			tryAgain = ""

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt = ClientRPC(rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)

			fmt.Print("Want to try again (y) -> yes | (n) -> no: ")
			fmt.Scan(&tryAgain)

			if tryAgain == "n" {
				break coonLoopRPC
			}
		}
	}

	// printPrimes(primes)
	// fmt.Print("RTT: ", rtt)
}

func ClientRPC(rng int, calcType string) ([]int, time.Duration) {
	// conecta ao servidor
	client, err := rpc.Dial("tcp", "localhost:1313")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	defer client.Close()

	var reply shared.Reply

	// invoca operação remota
	req := shared.Request{Rng: rng, Type: calcType}

	var startTime, endTime time.Time
	startTime = time.Now()

	err = client.Call("SieveCalcRPC.RpcBlockConcSieve", req, &reply)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	endTime = time.Now()

	return reply.Result, endTime.Sub(startTime)
}

func StartConnectionTCP() *net.TCPConn {
	// retorna o endereço do endpoint
	r, err := net.ResolveTCPAddr("tcp", "localhost:1314")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// connecta ao servidor (sem definir uma porta local)
	conn, err := net.DialTCP("tcp", nil, r)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	return conn
}

func SendMessageTCP(conn *net.TCPConn, rng int, calcType string) ([]int, time.Duration) {
	var response shared.Reply

	// cria enconder/decoder
	jsonDecoder := json.NewDecoder(conn)
	jsonEncoder := json.NewEncoder(conn)

	// prepara request
	msgToServer := shared.Request{Type: calcType, Rng: rng}

	var startTime, endTime time.Time
	startTime = time.Now()

	// serializa e envia request para o servidor
	err := jsonEncoder.Encode(msgToServer)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// recebe resposta do servidor
	err = jsonDecoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	endTime = time.Now()

	return response.Result, endTime.Sub(startTime)
}

func CloseConnectionTCP(conn *net.TCPConn) {
	err := conn.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func SieveClientTCP(rng int, calcType string) ([]int, time.Duration) {
	conn := StartConnectionTCP()
	defer CloseConnectionTCP(conn)

	return SendMessageTCP(conn, rng, calcType)
}

func StartConnectionUDP() *net.UDPConn {
	addr, err := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(shared.SievePort))
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	return conn
}

func SendMessageUDP(conn *net.UDPConn, rng int, calcType string) ([]int, time.Duration) {
	var response shared.Reply

	//decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	request := shared.Request{Type: calcType, Rng: rng}

	var startTime, endTime time.Time
	startTime = time.Now()

	err := encoder.Encode(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	var buffer [1024]byte
	var packetCount int

	n, _, err := conn.ReadFromUDP(buffer[:])
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(buffer[:n], &packetCount)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	var data = make([]byte, 0, (packetCount-1)*1024)

	for i := 0; i < packetCount; i++ {

		// Receive a chunk
		n, _, err := conn.ReadFromUDP(buffer[:])

		// err = decoder.Decode(&chunk)
		if err != nil {
			fmt.Println(err)
		}
		// Process the chunk (e.g., append it to the data)
		data = append(data, buffer[:n]...)

		// if i%100 == 0 {
		// 	fmt.Println(i)
		// }

	}

	//fmt.Println(data)

	err = json.Unmarshal(data, &response)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	endTime = time.Now()

	return response.Result, endTime.Sub(startTime)
}

func CloseConnectionUDP(conn *net.UDPConn) {
	err := conn.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func SieveClientUDP(rng int, calcType string) ([]int, time.Duration) {
	conn := StartConnectionUDP()
	defer CloseConnectionUDP(conn)

	return SendMessageUDP(conn, rng, calcType)
}

func printPrimes(primes []int) {
	length := len(primes)

	// imprime os primeiros e os ultimos 10 primos
	if length > 20 {
		fmt.Println("Found ", length, " primes:\n", primes[:10], " ... ", primes[length-10:])
	} else {
		fmt.Println("Found", length, "primes:\n", primes)
	}
}
