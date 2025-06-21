package config

import (
	"fmt"
	"go-saga-pattern/commoner/utils"
	"net"
	"os"
)

type ServerConfig struct {
	UserHTTPAddr string
	UserHTTPPort string
	UserSvcName  string

	UserGRPCAddr         string
	UserGRPCPort         string
	UserGRPCInternalAddr string

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
		UserHTTPAddr: utils.GetEnv("USER_HTTP_ADDR"),
		UserHTTPPort: utils.GetEnv("USER_HTTP_PORT"),
		UserSvcName:  utils.GetEnv("USER_SVC_NAME"),

		UserGRPCAddr:         utils.GetEnv("USER_GRPC_ADDR"),
		UserGRPCPort:         utils.GetEnv("USER_GRPC_PORT"),
		UserGRPCInternalAddr: ip,

		ConsulAddr: consulAddr,
	}
}
