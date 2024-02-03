package client

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
		conn := startConnectionUDP()

	connLoopUDP:
		for {
			calcType = ""
			for calcType != "seq" && calcType != "conc" && calcType != "blk_conc" {
				fmt.Print("Choose (seq) -> sequential | (conc) -> concurrent | (blk_conc) -> block_concurrent | (q) -> quit: ")
				fmt.Scan(&calcType)

				if calcType == "q" {
					closeConnectionUDP(conn)
					break connLoopUDP
				}
			}

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt = sendMessageUDP(conn, rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)
		}
	} else {
		conn := startConnectionTCP()

	connLoopTCP:
		for {
			calcType = ""
			for calcType != "seq" && calcType != "conc" && calcType != "blk_conc" {
				fmt.Print("Choose (seq) -> sequential | (conc) -> concurrent | (blk_conc) -> block_concurrent | (q) -> quit: ")
				fmt.Scan(&calcType)

				if calcType == "q" {
					closeConnectionTCP(conn)
					break connLoopTCP
				}
			}

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt = sendMessageTCP(conn, rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)
		}
	}

	// printPrimes(primes)
	// fmt.Print("RTT: ", rtt)
}

func startConnectionTCP() *net.TCPConn {
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

func sendMessageTCP(conn *net.TCPConn, rng int, calcType string) ([]int, time.Duration) {
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

func closeConnectionTCP(conn *net.TCPConn) {
	err := conn.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func SieveClientTCP(rng int, calcType string) ([]int, time.Duration) {
	conn := startConnectionTCP()
	defer closeConnectionTCP(conn)

	return sendMessageTCP(conn, rng, calcType)
}

func startConnectionUDP() *net.UDPConn {
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

func sendMessageUDP(conn *net.UDPConn, rng int, calcType string) ([]int, time.Duration) {
	var response shared.Reply

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	request := shared.Request{Type: calcType, Rng: rng}

	var startTime, endTime time.Time
	startTime = time.Now()

	err := encoder.Encode(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	err = decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	endTime = time.Now()

	return response.Result, endTime.Sub(startTime)
}

func closeConnectionUDP(conn *net.UDPConn) {
	err := conn.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func SieveClientUDP(rng int, calcType string) ([]int, time.Duration) {
	conn := startConnectionUDP()
	defer closeConnectionUDP(conn)

	return sendMessageUDP(conn, rng, calcType)
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
