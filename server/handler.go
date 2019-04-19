package server

import (
	"context"

	"github.com/abferm/remoteI2C/gen-go/rpc"
	"github.com/juju/loggo"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

var logger = loggo.GetLogger("i2c-server")

type I2CHandler struct {
	bus i2c.Bus
}

func NewI2CHandler(bus i2c.Bus) *I2CHandler {
	logger.SetLogLevel(loggo.TRACE)
	return &I2CHandler{bus: bus}
}

func (h *I2CHandler) String(ctx context.Context) (r string, err error) {
	r = "remote:" + h.bus.String()
	logger.Tracef("Got Call to String() -> (%q)", r)
	return r, nil
}

// Parameters:
//  - Addr
//  - W
//  - Length
func (h *I2CHandler) Tx(ctx context.Context, addr int16, w []byte, length int32) (r []byte, err error) {
	r = make([]byte, int(length))
	err = h.bus.Tx(uint16(addr), w, r)
	if err != nil {
		err = convertError(err)
		return []byte{}, err
	}
	logger.Tracef("Got call to Tx(%d, %2X, %d) -> (%2X)", uint16(addr), w, length, r)
	return r, nil
}

// Parameters:
//  - MicroHertz
func (h *I2CHandler) SetSpeed(ctx context.Context, microHertz int64) (err error) {
	logger.Tracef("Got call to SetSpeed(%s)", physic.Frequency(microHertz).String())
	err = h.bus.SetSpeed(physic.Frequency(microHertz))
	if err != nil {
		return convertError(err)
	}
	return nil
}

func convertError(err error) (err1 *rpc.Error) {
	if err == nil {
		return nil
	}
	err1 = rpc.NewError()
	err1.Message = err.Error()
	return err1
}
