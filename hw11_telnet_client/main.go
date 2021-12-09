package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var timeout time.Duration

func init() {
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout (ns)")
}

var tlnLogger *log.Logger

func main() {
	tlnLogger = log.New(os.Stderr, "*** INFO: ", 0)

	flag.Parse()
	var host string
	var port string

	switch len(os.Args) {
	case 3:
		host = os.Args[1]
		port = os.Args[2]
	case 4:
		host = os.Args[2]
		port = os.Args[3]
	default:
		tlnLogger.Fatal("*** incorrect number of input parameters")
	}

	client := NewTelnetClient(net.JoinHostPort(host, port), timeout, os.Stdin, os.Stdout)

	tlnLogger.Printf("connecting to %s %s", host, port)
	if err := client.Connect(); err != nil {
		tlnLogger.Fatal("failed to connect " + err.Error())
	}
	tlnLogger.Printf("connected to %s %s  ", host, port)

	run(client)

	tlnLogger.Println("done")
}

func run(client TelnetClient) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGINT)

	sendCh := createRoutine(client.Send)

	go func() {
		defer client.Close()
		select {
		case <-terminate:
			tlnLogger.Println("interrupted")
		case <-sendCh:
			tlnLogger.Println("failed to send data")
		}
	}()

	<-createRoutine(client.Receive)
	<-sendCh
}

func createRoutine(f func() error) <-chan error {
	errChain := make(chan error)
	go func() {
		defer close(errChain)
		errChain <- f()
	}()
	return errChain
}
