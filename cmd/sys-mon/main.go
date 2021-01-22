package main

import (
	"flag"
	"fmt"
	//	"github.com/hrapovd1/final_project/pkg/sysmon"
)

func main() {
	// Program flags
	isDaemon := flag.Bool("d", false, "daemon mode")
	isHelp := flag.Bool("h", false, "help output")
	dataBuff := flag.Uint("b", 5, "average data period in sec")
	port := flag.Uint("p", 8080, "server default tcp port")
	answPeriod := flag.Uint("a", 5, "server answer period in sec")
	flag.Parse()

	if *isHelp {
		fmt.Println("sys-mon [-d] [-b n] [-a m] [-p PORT]\n\t-d background mode\n\t-p PORT server port to listen clients\n\t-b n data buffer in seconds\n\t-a m answer period to client every m seconds\n\t-h this help")
		return
	}
	if *isDaemon {
		fmt.Println("Program will be run in background.")
	} else {
		fmt.Println("Program will be run in foreground.")
	}
	if *port < 1 {
		fmt.Println("port must be from 1 to 65535")
	} else {
		fmt.Printf("Server open %v port\n", *port)
	}
	if *dataBuff > 0 {
		fmt.Printf("wait %v seconds to get statistics\n", *dataBuff)
	}
	if *answPeriod > 0 {
		fmt.Printf("next answer through %v seconds\n", *answPeriod)
	}
}
