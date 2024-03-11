package main

import (
	"exercicio5/client"
	"exercicio5/server"
	"fmt"
)

func main(){
	var chosen string

	for chosen != "s" && chosen != "c" && chosen != "g" {
		fmt.Print("Choose (s) -> server | (c) -> client | (g) -> graph: ")
		fmt.Scan(&chosen)
	}

	if chosen == "s" {
		server.Run()
	} else if chosen == "c" {
		client.Run()
	} else {
		// graph.Run()
	}
}