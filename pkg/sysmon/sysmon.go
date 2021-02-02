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

	/*
	   type All struct {
	   	state         protoimpl.MessageState
	   	sizeCache     protoimpl.SizeCache
	   	unknownFields protoimpl.UnknownFields

	   	LoadAverage  *LoadAverage      `protobuf:"bytes,1,opt,name=loadAverage,proto3" json:"loadAverage,omitempty"`
	   	Cpu          *Cpu              `protobuf:"bytes,2,opt,name=cpu,proto3" json:"cpu,omitempty"`
	   	Disk         *Disk             `protobuf:"bytes,3,opt,name=disk,proto3" json:"disk,omitempty"`
	   	Partitions   []*Fs             `protobuf:"bytes,4,rep,name=partitions,proto3" json:"partitions,omitempty"`
	   	Connections  *TcpConnections   `protobuf:"bytes,5,opt,name=connections,proto3" json:"connections,omitempty"`
	   	Listners     []*Listen         `protobuf:"bytes,6,rep,name=listners,proto3" json:"listners,omitempty"`
	   	ProtoTalkers []*NetProtoTalker `protobuf:"bytes,7,rep,name=protoTalkers,proto3" json:"protoTalkers,omitempty"`
	   	RateTalker   []*NetRateTalker  `protobuf:"bytes,8,rep,name=rateTalker,proto3" json:"rateTalker,omitempty"`
	   }
	*/
	all := smgrpc.All{
		LoadAverage: &smgrpc.LoadAverage{Load: 5},
		Cpu:         &smgrpc.Cpu{Sys: 5, User: 30, Idle: 65},
		Disk:        &smgrpc.Disk{Tps: 7, Kbps: 100},
		Connections: &smgrpc.TcpConnections{Count: 10},
		Partitions: []*smgrpc.Fs{
			{Name: "/", Used: 30, Iused: 1},
			{Name: "/home", Used: 10, Iused: 5},
		},
		Listners: []*smgrpc.Listen{
			{Cmd: "bind", User: "nobody", Pid: 456, Proto: "Tcp", Port: 53},
			{Cmd: "sys-mon", User: "dima", Pid: 7654, Proto: "Tcp", Port: 8080},
		},
		ProtoTalkers: []*smgrpc.NetProtoTalker{
			{Proto: "Tcp", Bytes: 456789, Rate: 100},
			{Proto: "ICMP", Bytes: 12345, Rate: 50},
		},
		RateTalker: []*smgrpc.NetRateTalker{
			{Proto: "Tcp", Sport: 3456, Dport: 80, Bps: 30},
			{Proto: "Tcp", Sport: 4321, Dport: 80, Bps: 28},
		},
	}

	err := out.Send(&all)
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
