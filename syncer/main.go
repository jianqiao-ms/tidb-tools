// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/juju/errors"
	"github.com/ngaut/log"
)

func main() {
	cfg := NewConfig()
	err := cfg.Parse(os.Args[1:])
	switch errors.Cause(err) {
	case nil:
	case flag.ErrHelp:
		os.Exit(0)
	default:
		log.Errorf("parse cmd flags err %s\n", err)
		os.Exit(2)
	}

	log.SetLevelByString(cfg.LogLevel)

	if len(cfg.LogFile) > 0 {
		log.SetOutputByName(cfg.LogFile)
		log.SetHighlighting(false)

		if cfg.LogRotate == "hour" {
			log.SetRotateByHour()
		} else {
			log.SetRotateByDay()
		}
	}

	log.Infof("%v", cfg)

	syncer := NewSyncer(cfg)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sc
		log.Infof("Got signal [%d] to exit.", sig)
		syncer.Close()
	}()

	go func() {
		err1 := http.ListenAndServe(cfg.PprofAddr, nil)
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	err = syncer.Start()
	if err != nil {
		log.Error(errors.ErrorStack(err))
	}
}
