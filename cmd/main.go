package main

import (
	"flag"
	"log"
	"os"

	"github.com/mister-turtle/flexitty/webserver"
)

func main() {

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	log.Println("FlexiTTY - Web socket based multiplayer TTY")

	argListen := flag.String("l", "127.0.0.1", "Address to listen on, default 127.0.0.1")
	argPort := flag.Int("p", 8000, "Port to bind to, default 8000")

	flag.Parse()

	options := webserver.Options{
		Address: *argListen,
		Port:    *argPort,
	}

	log.Printf("Starting server on %s:%d\n", options.Address, options.Port)
	log.Fatal(webserver.Start(options))
}
