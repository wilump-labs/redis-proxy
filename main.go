package main

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net"
)

const proxyPort = 16379

var (
	host string
	port string
)

func main() {
	initConfig()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", proxyPort))
	if err != nil {
		log.Fatalf("Failed to setup listener: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept listener: %v", err)
		}
		go forward(conn, host, port)
	}
}

func initConfig() {
	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	host = viper.GetString("host")
	port = viper.GetString("port")
}

func forward(conn net.Conn, host, port string) {
	destination := fmt.Sprintf("%s:%s", host, port)

	client, err := net.Dial("tcp", destination)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
		return
	}

	log.Printf("Connected to redis %s\n", destination)

	done := make(chan struct{})

	go func() {
		defer conn.Close()
		defer client.Close()
		io.Copy(client, conn)
		done <- struct{}{}
	}()

	go func() {
		defer conn.Close()
		defer client.Close()
		io.Copy(conn, client)
		done <- struct{}{}
	}()

	<-done
	<-done
}
