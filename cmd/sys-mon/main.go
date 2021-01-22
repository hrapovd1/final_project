package main

import (
	"flag"
	"fmt"

	"github.com/hrapovd1/final_project/pkg/sysmon"
)

func main() {
	// Program flags
	isDaemon := flag.Bool("d", false, "daemon mode")
	isHelp := flag.Bool("h", false, "help output")
	dataBuff := flag.Uint("b", 5, "average data period in sec")
	port := flag.Uint("p", 8080, "server default tcp port")
	answPeriod := flag.Uint("a", 5, "server answer period in sec")
	flag.Parse()

	if isDaemon {
		fmt.Println("Program will be run in background.")
	} else {
		fmt.Println("Program will be run in foregraund.")
	}
	if isHelp {
		fmt.Println("sys-mon [-d] [-b n] [-a m] -p PORT")
	}
	if port < 1 {
		fmt.Println("port must be from 1 to 65535")
	} else {
		fmt.Printf("Server open %v port", *port)
	}
	if dataBuff > 0 {
		fmt.Printf("wait %v seconds to get statistics", *dataBuff)
	}
	if answPeriod > 0 {
		fmt.Printf("next answer through %v seconds", *answPeriod)
	}
}
