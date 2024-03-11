package shared

import (
	"log"
	"math/rand"
)

const SievePort = 4040
const ResponseQueue = "response_queue"
const RequestQueue = "request_queue"

type Request struct {
	Type string
	Rng  int
}

type Reply struct {
	Result []int
}

func ErrCheck(err error, msg string) {
	if err != nil {
		log.Fatalf("%s!!: %s", msg, err)
	}
}

func RandomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(RandInt(65, 90))
	}
	return string(bytes)
}

func RandInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
