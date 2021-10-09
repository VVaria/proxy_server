package proxy

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type Proxy struct {
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		handleHTTPS(w, r)
	} else {
		handleHTTP(w, r)
	}
}
func copyHeader(to, from http.Header) {
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Request HTTP: ", r.Method, r.URL)
	r.Header.Del("Proxy-Connection")
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal("handleHTTP: ", err.Error())
		return
	}
	defer response.Body.Close()

	copyHeader(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}

func connectHijacker(w http.ResponseWriter) (net.Conn, error) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return nil, errors.New("hijacking not supported")
	}

	client, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	return client, nil
}

func sendData(d io.WriteCloser, s io.ReadCloser) {
	defer d.Close()
	defer s.Close()
	io.Copy(d, s)
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Println("Request HTTPS: ", r.Method, r.URL)

	destination, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal("handleHTTPS handleShake: ", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)


	source, err := connectHijacker(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal("connectHijacker:", err)
		return
	}

	go sendData(destination, source)
	go sendData(source, destination)
}