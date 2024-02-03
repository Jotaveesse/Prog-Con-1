package main

import (
	"encoding/json"
	"exercicio3/shared"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {

	var rng int
	var conn_type, calcType string

	for conn_type != "u" && conn_type != "t" {
		fmt.Print("Choose (u) -> udp | (t) -> tcp: ")
		fmt.Scan(&conn_type)
	}

	var primes []int
	var rtt time.Duration

	if conn_type == "u" {
		conn := StartConnectionUDP()

	connLoopUDP:
		for {
			calcType = ""
			for calcType != "seq" && calcType != "conc" && calcType != "blk_conc" {
				fmt.Print("Choose (seq) -> sequential | (conc) -> concurrent | (blk_conc) -> block_concurrent | (q) -> quit: ")
				fmt.Scan(&calcType)

				if calcType == "q" {
					CloseConnectionUDP(conn)
					break connLoopUDP
				}
			}

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt = SendMessageUDP(conn, rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)
		}
	} else {
		conn := StartConnectionTCP()

	connLoopTCP:
		for {
			calcType = ""
			for calcType != "seq" && calcType != "conc" && calcType != "blk_conc" {
				fmt.Print("Choose (seq) -> sequential | (conc) -> concurrent | (blk_conc) -> block_concurrent | (q) -> quit: ")
				fmt.Scan(&calcType)

				if calcType == "q" {
					CloseConnectionTCP(conn)
					break connLoopTCP
				}
			}

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt = SendMessageTCP(conn, rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)
		}
	}

	// printPrimes(primes)
	// fmt.Print("RTT: ", rtt)
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
