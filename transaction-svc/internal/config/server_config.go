package config

import (
	"fmt"
	"go-saga-pattern/commoner/utils"
	"net"
	"os"
)

type ServerConfig struct {
	TransactionHTTPAddr string
	TransactionHTTPPort string
	TransactionSvcName  string

	ListenerHTTPAddr string
	ListenerHTTPPort string
	ListenerSvcName  string

	TransactionGRPCAddr         string
	TransactionGRPCPort         string
	TransactionGRPCInternalAddr string

	ConsulAddr string
}

func NewServerConfig() *ServerConfig {
	hostname, _ := os.Hostname()
	addrs, _ := net.LookupIP(hostname)

	var ip string
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil && !ipv4.IsLoopback() {
			ip = ipv4.String()
			break
		}
	}

	consulAddr := fmt.Sprintf("%s:%s", utils.GetEnv("CONSUL_HOST"), utils.GetEnv("CONSUL_PORT"))

	return &ServerConfig{
		TransactionHTTPAddr: utils.GetEnv("TRANSACTION_HTTP_ADDR"),
		TransactionHTTPPort: utils.GetEnv("TRANSACTION_HTTP_PORT"),
		TransactionSvcName:  utils.GetEnv("TRANSACTION_SVC_NAME"),

		ListenerHTTPAddr: utils.GetEnv("LISTENER_HTTP_ADDR"),
		ListenerHTTPPort: utils.GetEnv("LISTENER_HTTP_PORT"),
		ListenerSvcName:  utils.GetEnv("LISTENER_SVC_NAME"),

		TransactionGRPCAddr:         utils.GetEnv("TRANSACTION_GRPC_ADDR"),
		TransactionGRPCPort:         utils.GetEnv("TRANSACTION_GRPC_PORT"),
		TransactionGRPCInternalAddr: ip,

		ConsulAddr: consulAddr,
	}
}
