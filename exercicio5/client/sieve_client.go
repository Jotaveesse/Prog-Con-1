package client

import (
	"encoding/json"
	"exercicio5/shared"
	"fmt"
	"github.com/streadway/amqp"
	"net/rpc"
	"os"
	"time"
)

func Run() {

	var rng int
	var conn_type, tryAgain, calcType string

	calcType = "blk_conc"

	for conn_type != "rp" && conn_type != "ra" {
		fmt.Print("Choose (rp) -> Go RPC | (ra) -> RabbitMQ: ")
		fmt.Scan(&conn_type)
	}

	var primes []int
	var rtt time.Duration

	if conn_type == "rp" {
		client := StartConnectionRPC()

	coonLoopRPC:
		for {
			tryAgain = ""

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt, _ = SendMessageRPC(client, rng, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)

			fmt.Print("Want to try again (y) -> yes | (n) -> no: ")
			fmt.Scan(&tryAgain)

			if tryAgain == "n" {
				CloseConnectionRPC(client)
				break coonLoopRPC
			}
		}
	} else {
		conn, ch, replyQueue, msgs := StartConnectionRabbitMQ()

	coonLoopRabbitMQ:
		for {
			tryAgain = ""

			fmt.Print("Choose the range: ")
			fmt.Scan(&rng)

			primes, rtt, _ = SendMessageRabbitMQ(rng, ch, replyQueue, msgs, calcType)

			printPrimes(primes)
			fmt.Println("RTT: ", rtt)

			fmt.Print("Want to try again (y) -> yes | (n) -> no: ")
			fmt.Scan(&tryAgain)

			if tryAgain == "n" {
				CloseConnectionRabbitMQ(conn, ch)
				break coonLoopRabbitMQ
			}
		}
	}
	// printPrimes(primes)
	// fmt.Print("RTT: ", rtt)
}

func StartConnectionRPC() *rpc.Client {
	client, err := rpc.Dial("tcp", "localhost:1313")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	return client
}

func SendMessageRPC(client *rpc.Client, rng int, calcType string) ([]int, time.Duration, time.Duration) {
	var reply shared.Reply

	var startTime, endTime time.Time
	// invoca operação remota
	req := shared.Request{Rng: rng, Type: calcType}

	startTime = time.Now()

	err := client.Call("SieveCalcRPC.RpcBlockConcSieve", req, &reply)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	endTime = time.Now()

	return reply.Result, endTime.Sub(startTime), reply.ProcessTime
}

func CloseConnectionRPC(client *rpc.Client) {
	err := client.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func StartConnectionRabbitMQ() (*amqp.Connection, *amqp.Channel, amqp.Queue, <-chan amqp.Delivery) {
	// conecta ao broker
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	shared.ErrCheck(err, "Não foi possível se conectar ao servidor de mensageria")

	// cria o canal
	ch, err := conn.Channel()
	shared.ErrCheck(err, "Não foi possível estabelecer um canal de comunicação com o servidor de mensageria")

	// declara a fila para as respostas
	replyQueue, err := ch.QueueDeclare(
		shared.ResponseQueue,
		false,
		false,
		true,
		false,
		nil,
	)
	shared.ErrCheck(err, "Falha ao declarar a fila de resposta")

	// cria servidor da fila de response
	msgs, err := ch.Consume(
		replyQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil)
	shared.ErrCheck(err, "Falha ao registrar o servidor no broker")

	return conn, ch, replyQueue, msgs
}

func SendMessageRabbitMQ(rng int, ch *amqp.Channel, replyQueue amqp.Queue, msgs <-chan amqp.Delivery, calcType string) ([]int, time.Duration, time.Duration) {
	// prepara mensagem
	msgRequest := shared.Request{Rng: rng, Type: calcType}
	msgRequestBytes, err := json.Marshal(msgRequest)
	shared.ErrCheck(err, "Falha ao serializar a mensagem")

	correlationID := shared.RandomString(32)

	// marca o tempo de inicio
	startTime := time.Now()

	err = ch.Publish(
		"",
		shared.RequestQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: correlationID,
			ReplyTo:       replyQueue.Name,
			Body:          msgRequestBytes,
		},
	)
	shared.ErrCheck(err, "Falha ao publicar a mensagem no broker")

	// recebe mensagem do servidor de mensageria
	m := <-msgs

	// deserializada e imprime mensagem na tela
	reply := shared.Reply{}
	err = json.Unmarshal(m.Body, &reply)
	shared.ErrCheck(err, "Erro na deserialização da resposta")

	// marca o tempo de fim
	endTime := time.Now()

	return reply.Result, endTime.Sub(startTime), reply.ProcessTime
}

func CloseConnectionRabbitMQ(conn *amqp.Connection, ch *amqp.Channel) {
	err := ch.Close()
	shared.ErrCheck(err, "Não foi possível fechar o canal de comunicação com o servidor de mensageria")

	err = conn.Close()
	shared.ErrCheck(err, "Não foi possível fechar a conexão com o servidor de mensageria")
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
