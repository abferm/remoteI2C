// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// i2c-io communicates to an I²C device.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/abferm/remoteI2C/client"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"

	"github.com/apache/thrift/lib/go/thrift"
)

func getRemoteBus(protocol *string, framed, buffered *bool, addr *string, secure *bool) (*client.RemoteBus, error) {
	var protocolFactory thrift.TProtocolFactory
	switch *protocol {
	case "compact":
		protocolFactory = thrift.NewTCompactProtocolFactory()
	case "simplejson":
		protocolFactory = thrift.NewTSimpleJSONProtocolFactory()
	case "json":
		protocolFactory = thrift.NewTJSONProtocolFactory()
	case "binary", "":
		protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	default:
		fmt.Fprint(os.Stderr, "Invalid protocol specified", protocol, "\n")
		flag.Usage()
		os.Exit(1)
	}

	var transportFactory thrift.TTransportFactory
	if *buffered {
		transportFactory = thrift.NewTBufferedTransportFactory(8192)
	} else {
		transportFactory = thrift.NewTTransportFactory()
	}

	if *framed {
		transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	}

	// always be client
	fmt.Printf("*secure = '%v'\n", *secure)
	fmt.Printf("*addr   = '%v'\n", *addr)
	return client.Dial(transportFactory, protocolFactory, *addr, *secure)
}

func mainImpl() error {
	addr := flag.Int("a", -1, "I²C device address to query")
	protocol := flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
	framed := flag.Bool("framed", false, "Use framed transport")
	buffered := flag.Bool("buffered", false, "Use buffered transport")
	serverAddr := flag.String("server", "localhost:9090", "Address of server")
	secure := flag.Bool("secure", false, "Use tls secure transport")
	verbose := flag.Bool("v", false, "verbose mode")
	// TODO(maruel): This is not generic enough.
	write := flag.Bool("w", false, "write instead of reading")
	reg := flag.Int("r", -1, "register to address")
	var hz physic.Frequency
	flag.Var(&hz, "hz", "I²C bus speed (may require root)")
	l := flag.Int("l", 1, "length of data to read; ignored if -w is specified")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

	if *addr < 0 || *addr >= 1<<9 {
		return fmt.Errorf("-a is required and must be between 0 and %d", 1<<9-1)
	}
	if *reg < 0 || *reg > 255 {
		return errors.New("-r must be between 0 and 255")
	}
	if *l <= 0 || *l > 255 {
		return errors.New("-l must be between 1 and 255")
	}

	bus, err := getRemoteBus(protocol, framed, buffered, serverAddr, secure)
	if err != nil {
		return err
	}
	defer bus.Close()

	var buf []byte
	if *write {
		if flag.NArg() == 0 {
			return errors.New("specify data to write as a list of hex encoded bytes")
		}
		buf = make([]byte, 1, flag.NArg()+1)
		buf[0] = byte(*reg)
		for _, a := range flag.Args() {
			b, err := strconv.ParseUint(a, 0, 8)
			if err != nil {
				return err
			}
			buf = append(buf, byte(b))
		}
	} else {
		if flag.NArg() != 0 {
			return errors.New("do not specify bytes when reading")
		}
		buf = make([]byte, *l)
	}

	if hz != 0 {
		if err := bus.SetSpeed(hz); err != nil {
			return err
		}
	}

	d := i2c.Dev{Bus: bus, Addr: uint16(*addr)}
	if *write {
		_, err = d.Write(buf)
	} else {
		if err = d.Tx([]byte{byte(*reg)}, buf); err != nil {
			return err
		}
		for i, b := range buf {
			if i != 0 {
				if _, err = fmt.Print(", "); err != nil {
					break
				}
			}
			if _, err = fmt.Printf("0x%02X", b); err != nil {
				break
			}
		}
		_, err = fmt.Print("\n")
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "i2c-io: %s.\n", err)
		os.Exit(1)
	}
}
