package sysmon

import (
	"log"
	"net"
	"os"
	"time"

	smgrpc "github.com/hrapovd1/final_project/pkg/smgrpc"
	grpc "google.golang.org/grpc"
)

type Sysmon interface {
	Run(logger log.Logger, sysCh <-chan os.Signal) error
}

type sysmonState struct {
	port       uint
	dataBuff   uint
	answPeriod uint
}

var smState *sysmonState

type listner struct {
	cmd   string
	pid   uint
	user  string
	proto uint
	port  uint
}

type monState struct {
	loadAver   uint
	cpu        [3]uint
	diskLoad   [2]uint
	fsUsage    map[string][2]uint
	netListner listner
	netSocks   uint
}

func NewSysmon(dataBuff, answPeriod, port uint) Sysmon {
	smState = &sysmonState{
		port:       port,
		dataBuff:   dataBuff,
		answPeriod: answPeriod,
	}
	return &monState{
		loadAver:   0,
		cpu:        [...]uint{0, 0, 0},
		diskLoad:   [...]uint{0, 0},
		fsUsage:    make(map[string][2]uint, 1),
		netListner: listner{},
		netSocks:   0,
	}
}

func (mst *monState) Run(logger log.Logger, sysCh <-chan os.Signal) error {
	srvSock := "0.0.0.0" + ":" + string(smState.port)
	lsn, err := net.Listen("tcp", srvSock)
	if err != nil {
		logger.Fatal(err)
	}

	server := grpc.NewServer()

	return nil
}
