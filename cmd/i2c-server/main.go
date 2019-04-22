package main

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"flag"
	"fmt"
	"os"

	"github.com/juju/loggo"

	"periph.io/x/periph/conn/i2c/i2creg"

	"github.com/abferm/remoteI2C/server"

	"github.com/apache/thrift/lib/go/thrift"
	"periph.io/x/periph/host"
)

const defaultLogLevel = loggo.WARNING

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	busName := flag.String("b", "", "I²C bus to use")
	protocol := flag.String("P", "binary", "Specify the protocol (binary, compact, json, simplejson)")
	framed := flag.Bool("framed", false, "Use framed transport")
	buffered := flag.Bool("buffered", false, "Use buffered transport")
	addr := flag.String("addr", "localhost:9090", "Address to listen to")
	secure := flag.Bool("secure", false, "Use tls secure transport")
	logLevel := flag.String("log-level", defaultLogLevel.String(), "Logging verbosity level, one of: CRITICAL, ERROR, WARN(ING), INFO, DEBUG, TRACE")

	flag.Parse()

	logger := loggo.GetLogger("")
	loggoLevel, ok := loggo.ParseLevel(*logLevel)
	if !ok {
		logger.Warningf("Invalid log level %q, using default %q", *logLevel, defaultLogLevel.String())
		loggoLevel = defaultLogLevel
	}
	logger.SetLogLevel(loggoLevel)

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
		logger.Criticalf("Invalid protocol specified %q", *protocol)
		Usage()
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

	_, err := host.Init()
	if err != nil {
		panic(err)
	}

	bus, err := i2creg.Open(*busName)
	if err != nil {
		panic(err)
	}

	defer bus.Close()

	// always run server here
	if err := server.RunServer(bus, transportFactory, protocolFactory, *addr, *secure); err != nil {
		logger.Criticalf("error running server: %s", err.Error())
	}
}
