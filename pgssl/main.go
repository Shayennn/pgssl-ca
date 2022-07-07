package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

var options struct {
	listenAddress string
	pgAddress     string
	caCertPath    string
}

func argFatal(s string) {
	fmt.Fprintln(os.Stderr, s)
	flag.Usage()
	os.Exit(1)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage:  %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	// parse arguments
	flag.StringVar(&options.listenAddress, "l", "127.0.0.1:15432", "Listen address")
	flag.StringVar(&options.pgAddress, "p", "", "Postgres address")
	flag.StringVar(&options.caCertPath, "c", "", "caCertPath")
	flag.Parse()

	if options.pgAddress == "" {
		argFatal("postgres address must be specified")
	}
	if options.caCertPath == "" {
		argFatal("caCertPath must be specified")
	}

	// load client certificate and key
	caCertPEM, err := ioutil.ReadFile(options.caCertPath)
	if err != nil {
		panic("failed to read root certificate")
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		panic("failed to parse root certificate")
	}

	// create pgSSL instance
	pgSSL := &PgSSL{
		pgAddr:    options.pgAddress,
		clientCAs: roots,
	}

	// bind listening socket
	ln, err := net.Listen("tcp", options.listenAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on", ln.Addr())

	// start accepting connection
	var connNum int
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		connNum++
		log.Printf("[%3d] Accepted connection from %s\n", connNum, conn.RemoteAddr())

		// handle connection in goroutine
		go func(n int) {
			err := pgSSL.HandleConn(conn)
			if err != nil {
				log.Printf("[%3d] error in connection: %s", n, err)
			}
			log.Printf("[%3d] Closed connection from %s", n, conn.RemoteAddr())
		}(connNum)
	}
}
