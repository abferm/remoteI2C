package client

import (
	"crypto/tls"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"
)

func DialDefaults(addr string) (bus *RemoteBus, err error) {
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory := thrift.NewTTransportFactory()
	return Dial(transportFactory, protocolFactory, addr, false)
}

func Dial(transportFactory thrift.TTransportFactory, protocolFactory thrift.TProtocolFactory, addr string, secure bool) (bus *RemoteBus, err error) {
	var transport thrift.TTransport
	if secure {
		cfg := new(tls.Config)
		cfg.InsecureSkipVerify = true
		transport, err = thrift.NewTSSLSocket(addr, cfg)
	} else {
		transport, err = thrift.NewTSocket(addr)
	}
	if err != nil {
		fmt.Println("Error opening socket:", err)
		return nil, err
	}
	if transport == nil {
		return nil, fmt.Errorf("Error opening socket, got nil transport. Is server available?")
	}
	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		fmt.Println("Error getting transport:", err)
		return nil, err
	}
	if transport == nil {
		return nil, fmt.Errorf("Error from transportFactory.GetTransport(), got nil transport. Is server available?")
	}

	err = transport.Open()
	if err != nil {
		return nil, err
	}
	return NewRemoteBus(transport, protocolFactory), nil
}
