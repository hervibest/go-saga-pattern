package config

import (
	"fmt"
	"go-saga-pattern/commoner/utils"
	"net"
	"os"
)

type ServerConfig struct {
	ProductHTTPAddr string
	ProductHTTPPort string
	ProductSvcName  string

	ProductGRPCAddr         string
	ProductGRPCPort         string
	ProductGRPCInternalAddr string

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
		ProductHTTPAddr: utils.GetEnv("PRODUCT_HTTP_ADDR"),
		ProductHTTPPort: utils.GetEnv("PRODUCT_HTTP_PORT"),
		ProductSvcName:  utils.GetEnv("PRODUCT_SVC_NAME"),

		ProductGRPCAddr:         utils.GetEnv("PRODUCT_GRPC_ADDR"),
		ProductGRPCPort:         utils.GetEnv("PRODUCT_GRPC_PORT"),
		ProductGRPCInternalAddr: ip,

		ConsulAddr: consulAddr,
	}
}
