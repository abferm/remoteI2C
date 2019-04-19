package client

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"

	"periph.io/x/periph/conn/physic"

	"github.com/abferm/remoteI2C/gen-go/rpc"
)

type RemoteBus struct {
	rpcClient rpc.I2C
	t         thrift.TTransport
}

func NewRemoteBus(t thrift.TTransport, f thrift.TProtocolFactory) *RemoteBus {
	client := rpc.NewI2CClientFactory(t, f)
	return &RemoteBus{rpcClient: client, t: t}
}

// Implement i2c.Bus
func (remote *RemoteBus) String() string {
	r, _ := remote.rpcClient.String(context.Background())
	return r
}

// Tx does a transaction at the specified device address.
//
// Write is done first, then read. One of 'w' or 'r' can be omitted for a
// unidirectional operation.
func (remote *RemoteBus) Tx(addr uint16, w, r []byte) error {
	r1, err := remote.rpcClient.Tx(context.Background(), int16(addr), w, int32(len(r)))
	copy(r, r1)
	return err
}

// SetSpeed changes the bus speed, if supported.
//
// On linux due to the way the I²C sysfs driver is exposed in userland,
// calling this function will likely affect *all* I²C buses on the host.
func (remote *RemoteBus) SetSpeed(f physic.Frequency) error {
	return remote.rpcClient.SetSpeed(context.Background(), int64(f))
}

// Close connection to remote
func (remote *RemoteBus) Close() error {
	return remote.t.Close()
}
