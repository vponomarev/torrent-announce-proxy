package main

import (
	"crypto/tls"
	"flag"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

var (
	ConfigFileName = flag.String("config", "config.yml", "Config file name")
	config         Config
)

func main() {
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	config, err := loadConfig(*ConfigFileName)
	if err != nil {
		log.Fatal("Error loading config file [", ConfigFileName, "]:", err)
	}

	//	ctx := context.Background()
	cx := HTTPProcessor{config: &config}

	server := &http.Server{
		Addr:    config.HTTP.Listen,
		Handler: http.HandlerFunc(cx.requestHandler),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Fatal(server.ListenAndServe())
	//log.Fatal(server.ListenAndServeTLS(pemPath, keyPath))
}
