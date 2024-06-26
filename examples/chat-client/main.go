package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/danielpfeifer02/priority-moqtransport/examples/chat"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "address to connect to")
	flag.Parse()

	var c *chat.Client
	var err error
	ctx := context.Background()

	c, err = chat.NewQUICClient(ctx, *addr)

	if err != nil {
		log.Fatal(err)
	}
	if err := c.Run(); err != nil {
		fmt.Printf("run returned err: %v\n", err)
	}
	fmt.Println("Bye")
}
