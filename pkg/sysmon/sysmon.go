package sysmon

import (
	"fmt"
	"log"
	"net"
	"strconv"

	smgrpc "github.com/hrapovd1/final_project/pkg/smgrpc"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

type statServer struct {
	smgrpc.UnimplementedStatServer
}

func (sS *statServer) GetAll(query *smgrpc.Request, out smgrpc.Stat_GetAllServer) error {
	header := metadata.New(map[string]string{"application": "System monitoring"})
	out.SendHeader(header)

	fmt.Printf("request received: %v\n", query)

	err := out.Send(&smgrpc.All{})
	if err != nil {
		return err
	}
	return nil
}

type Sysmon interface {
	Run(doneCh <-chan interface{}, logger log.Logger) error
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

func (mst *monState) Run(doneCh <-chan interface{}, logger log.Logger) error {
	srvSock := ":" + strconv.Itoa(int(smState.port))
	lsn, err := net.Listen("tcp", srvSock)
	if err != nil {
		logger.Fatal(err)
	}

	server := grpc.NewServer()
	smgrpc.RegisterStatServer(server, &statServer{})

	logger.Printf("open port %v", smState.port)

	go func() {
		err = server.Serve(lsn)
		if err != nil {
			logger.Fatal("Unable to start server:", err)
		}
	}()

	go func() {
		<-doneCh
		server.Stop()
		defer lsn.Close()
	}()
	return nil
}
