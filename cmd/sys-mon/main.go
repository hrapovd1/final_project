package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	sysmon "github.com/hrapovd1/final_project/pkg/sysmon"
)

func main() {
	// Program flags
	dataBuff := flag.Uint("b", 5, "average data period in sec")
	port := flag.Uint("p", 8080, "server default tcp port")
	answPeriod := flag.Uint("a", 5, "server answer period in sec")

	flag.Parse()

	// Subscribe os signals
	sysSigCh := make(chan os.Signal, 1)
	signal.Notify(sysSigCh, syscall.SIGINT, syscall.SIGTERM)

	// Setup logger to stdout
	stdoutLog := log.New(os.Stdout, "", log.LstdFlags)

	// open done chanel
	doneCh := make(chan interface{})

	// Create and run sys-mon
	monInstance := sysmon.NewSysmon(*dataBuff, *answPeriod, *port)
	err := monInstance.Run(doneCh, stdoutLog)
	if err != nil {
		stdoutLog.Printf("ERROR: %v", err)
	}

	// Wait stop signal
	<-sysSigCh
	close(doneCh)
	stdoutLog.Println("Got stop signal")
}
