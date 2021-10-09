package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/VVaria/proxy_server/proxy"
)

func main() {
	p := &proxy.Proxy{}

	var pem, key, protocol string
	flag.StringVar(&pem, "pem", "cert.pem", "")
	flag.StringVar(&key, "key", "key.pem", "")
	flag.StringVar(&protocol, "proto", "http", "")
	flag.Parse()

	log.Println("Start server port 8080")
	server := &http.Server{
		Addr: ":8080",
		Handler: p,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	if protocol == "http" {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		err := server.ListenAndServeTLS(pem, key)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	}
}