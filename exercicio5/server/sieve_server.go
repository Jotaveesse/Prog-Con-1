package server

import (
	"exercicio5/server/service"
	"exercicio5/shared"
	"github.com/streadway/amqp"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"encoding/json"
)

func Run() {
	var conn_type string

	for conn_type != "rp" && conn_type != "ra" {
		fmt.Print("Choose (rp) -> Go RPC | (ra) -> RabbitMQ: ")
		fmt.Scan(&conn_type)
	}

	if conn_type == "rp" {
		SieveServerRPC()
	} else {
		SieveServerRabbitMQ()
	}

	fmt.Scanln()
}

func SieveServerRPC() {
	// cria uma instância da Crivo de crivo
	sieveCalculator := new(service.SieveCalcRPC)

	// cria um novo servidor RPC e registra a Crivo de crivo
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

func SieveServerRabbitMQ() {
	// cria conexão com o broker
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	shared.ErrCheck(err, "Não foi possível se conectar ao broker")
	defer conn.Close()

	// cria um canal
	ch, err := conn.Channel()
	shared.ErrCheck(err, "Não foi possível estabelecer um canal de comunicação com o broker")
	defer ch.Close()

	// declara a fila
	q, err := ch.QueueDeclare(
		shared.RequestQueue,
		false,
		false,
		false,
		false,
		nil)
	shared.ErrCheck(err, "Não foi possível criar a fila no broker")

	// prepara o recebimento de mensagens do cliente
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil)
	shared.ErrCheck(err, "Falha ao registrar o consumidor no broker")

	fmt.Println("Crivo pronto...")

	for d := range msgs {
		// recebe request
		msg := shared.Request{}
		err := json.Unmarshal(d.Body, &msg)
		shared.ErrCheck(err, "Falha ao desserializar a mensagem")

		// processa request
		r := service.SieveCalc{}.InvokeSieveCalc(msg)

		// prepara resposta
		replyMsg := shared.Reply{Result: r}
		replyMsgBytes, err := json.Marshal(replyMsg)
		shared.ErrCheck(err, "Falha ao serializar mensagem")

		// publica resposta
		err = ch.Publish(
			"",
			d.ReplyTo,
			false,
			false,
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: d.CorrelationId, // usa correlation id do request
				Body:          replyMsgBytes,
			},
		)
		shared.ErrCheck(err, "Falha ao enviar a mensagem para o broker")
	}


}
