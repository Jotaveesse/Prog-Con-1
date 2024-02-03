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

	for calcType != "seq" && calcType != "conc" && calcType != "blk_conc" {
		fmt.Print("Choose (seq) -> sequential | (conc) -> concurrent | (blk_conc) -> block_concurrent: ")
		fmt.Scan(&calcType)
	}

	fmt.Print("Choose the range: ")
	fmt.Scan(&rng)

	var primes []int
	var rtt time.Duration

	if conn_type == "u" {
		primes, rtt = SieveClientUDP(rng, calcType)
	} else {
		primes, rtt = SieveClientTCP(rng, calcType)
	}

	printPrimes(primes)
	fmt.Print("RTT: ", rtt)
}

func SieveClientTCP(rng int, calcType string) ([]int, time.Duration) {
	var response shared.Reply

	// retorna o endereço do endpoint
	r, err := net.ResolveTCPAddr("tcp", "localhost:1314")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	/// connecta ao servidor (sem definir uma porta local)
	conn, err := net.DialTCP("tcp", nil, r)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// fecha conexão
	defer func(conn *net.TCPConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// cria enconder/decoder
	jsonDecoder := json.NewDecoder(conn)
	jsonEncoder := json.NewEncoder(conn)

	// prepara request
	msgToServer := shared.Request{Type: calcType, Rng: rng}

	var startTime, endTime time.Time
	startTime = time.Now()

	// serializa e envia request para o servidor
	err = jsonEncoder.Encode(msgToServer)
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

func SieveClientUDP(rng int, calcType string) ([]int, time.Duration) {
	var response shared.Reply

	// resolve server address
	addr, err := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(shared.SievePort))
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// connect to server -- does not create a connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// create coder/decoder
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	// Close connection
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	// Create request
	request := shared.Request{Type: calcType, Rng: rng}

	var startTime, endTime time.Time
	startTime = time.Now()

	// Serialise and send request
	err = encoder.Encode(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	// Receive response from servidor
	err = decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	endTime = time.Now()

	return response.Result, endTime.Sub(startTime)
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
