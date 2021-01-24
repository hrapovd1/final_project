package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	sysmon "github.com/hrapovd1/final_project/pkg/sys-mon"
)

func main() {
	// Program flags
	isHelp := flag.Bool("h", false, "help output")
	dataBuff := flag.Uint("b", 5, "average data period in sec")
	port := flag.Uint("p", 8080, "server default tcp port")
	answPeriod := flag.Uint("a", 5, "server answer period in sec")
	flag.Parse()
	// Print help message and exit.
	if *isHelp {
		fmt.Println("sys-mon [-b n] [-a m] [-p PORT]\n\t-p PORT server port to listen clients\n\t-b n data buffer in seconds\n\t-a m answer period to client every m seconds\n\t-h this help")
		return
	}

	// Subscribe os signals
	sysSigCh := make(chan os.Signal, 1)
	signal.Notify(sysSigCh, syscall.SIGINT, syscall.SIGTERM)
	// Setup logger to stdout
	stdoutLog := log.New(os.Stdout, "", log.LstdFlags)

	// Create new sys-mon
	monInstance := sysmon.NewSysmon(*dataBuff, *answPeriod, *port)
	err := monInstance.Run(*stdoutLog, sysSigCh)
	if err != nil {
		stdoutLog.Printf("ERROR: %v", err)
	}

	stdoutLog.Println("wait for init service...")

	<-sysSigCh
	stdoutLog.Println("Got stop signal")
	syscall.Exit(0)
}
