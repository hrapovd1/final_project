package sysmon

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	smgrpc "github.com/hrapovd1/final_project/pkg/smgrpc"
	//smgrpc "../smgrpc"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

// Internal state of sys-mon process
type sysmonState struct {
	port       uint
	dataBuff   uint
	answPeriod uint
}

var smState *sysmonState

// Type for store OS network listner
type listner struct {
	cmd   string
	pid   uint
	user  string
	proto uint
	port  uint
}

// Type for store monitoring data scrape
type monState struct {
	loadAver   float32
	cpu        [3]float32
	diskLoad   map[string][]float32
	fsUsage    map[string][2]float32
	netListner listner
	netSocks   uint
}

type monStateBuff []monState

// System monitoring process
type Sysmon interface {
	Run(doneCh <-chan interface{}, logger *log.Logger) error
}

// System monitoring constructor
func NewSysmon(dataBuff, answPeriod, port uint) Sysmon {
	smState = &sysmonState{
		port:       port,
		dataBuff:   dataBuff,
		answPeriod: answPeriod,
	}
	smBuff := make(monStateBuff, dataBuff) // sys-mon scrapes buffer

	return &smBuff
}

// aggregate data from agents and sent "All" messages in grpc
func (mst *monStateBuff) runAggregate(doneCh <-chan interface{}, messages chan *smgrpc.All, logger *log.Logger) error {
	count := 0
	answer := time.NewTicker(time.Duration(smState.answPeriod) * time.Second)
	second := time.NewTicker(time.Second)
	condMu := new(sync.Mutex)
	cond := sync.NewCond(condMu)

	agents, err := runAgents(doneCh, cond, logger) //Function defined in agents.go file
	if err != nil {
		logger.Fatal("Run agents error: ", err)
	}

	// ask agents every second
	go func() {
		for {
			//TODO: add time-out reset
			select {
			case <-second.C:
				if count == len(*mst) {
					count = 0
				}
				var wg sync.WaitGroup
				wg.Add(6)
				cond.Broadcast()
				go func() {
					msg := <-agents["loadAver"]
					(*mst)[count].loadAver = msg.loadAver
					wg.Done()
				}()
				go func() {
					msg := <-agents["cpu"]
					(*mst)[count].cpu = msg.cpu
					wg.Done()
				}()
				go func() {
					msg := <-agents["diskLoad"]
					(*mst)[count].diskLoad = msg.diskLoad
					wg.Done()
				}()
				go func() {
					msg := <-agents["fsUsage"]
					(*mst)[count].fsUsage = msg.fsUsage
					wg.Done()
				}()
				go func() {
					msg := <-agents["netListner"]
					(*mst)[count].netListner = msg.netListner
					wg.Done()
				}()
				go func() {
					msg := <-agents["netSocks"]
					(*mst)[count].netSocks = msg.netSocks
					wg.Done()
				}()
				wg.Wait()
				count++
			case <-doneCh:
				second.Stop()
				return
			}
		}
	}()

	// fill "All" struct from monState buffer and sent to grpc
	go func() {
		for {
			select {
			case <-answer.C:
				for _, scrape := range *mst {
					/*
						    type All struct {
						   	state         protoimpl.MessageState
							sizeCache     protoimpl.SizeCache
							unknownFields protoimpl.UnknownFields

							LoadAverage  *LoadAverage      `protobuf:"bytes,1,opt,name=loadAverage,proto3" json:"loadAverage,omitempty"`
							Cpu          *Cpu              `protobuf:"bytes,2,opt,name=cpu,proto3" json:"cpu,omitempty"`
							Disk         []*Disk           `protobuf:"bytes,3,rep,name=disk,proto3" json:"disk,omitempty"`
							Partitions   []*Fs             `protobuf:"bytes,4,rep,name=partitions,proto3" json:"partitions,omitempty"`
							Connections  *TcpConnections   `protobuf:"bytes,5,opt,name=connections,proto3" json:"connections,omitempty"`
							Listners     []*Listen         `protobuf:"bytes,6,rep,name=listners,proto3" json:"listners,omitempty"`
							ProtoTalkers []*NetProtoTalker `protobuf:"bytes,7,rep,name=protoTalkers,proto3" json:"protoTalkers,omitempty"`
							RateTalker   []*NetRateTalker  `protobuf:"bytes,8,rep,name=rateTalker,proto3" json:"rateTalker,omitempty"`
					*/
					disksCount := len(scrape.diskLoad)
					disksLoad := make([]*smgrpc.Disk, disksCount)
					for disk, stats := range scrape.diskLoad {
						disksLoad[disksCount-1] = &smgrpc.Disk{Name: disk, Tps: stats[0], Kbps: stats[1]}
						disksCount--
					}
					messages <- &smgrpc.All{
						LoadAverage: &smgrpc.LoadAverage{Load: scrape.loadAver},
						Cpu:         &smgrpc.Cpu{Sys: scrape.cpu[0], User: scrape.cpu[1], Idle: scrape.cpu[2]},
						Disk:        disksLoad,
						Connections: &smgrpc.TcpConnections{Count: 10},
						Partitions: []*smgrpc.Fs{
							{Name: "/", Used: 30, Iused: 0},
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
				}
			case <-doneCh:
				answer.Stop()
				close(messages)
				return
			}
		}
	}()
	return nil
}

// grpc server implementation
type statServer struct {
	smgrpc.UnimplementedStatServer
}

// chanel for send scrapes
var allCh = make(chan *smgrpc.All)

func (sS *statServer) GetAll(query *smgrpc.Request, out smgrpc.Stat_GetAllServer) error {
	header := metadata.New(map[string]string{"application": "System monitoring", "timestamp": time.Now().String()})
	out.SendHeader(header)

	fmt.Printf("request received: %v\n", query)

	for msg := range allCh {
		err := out.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// sys-mon implementation
func (mst *monStateBuff) Run(doneCh <-chan interface{}, logger *log.Logger) error {
	err := mst.runAggregate(doneCh, allCh, logger)
	if err != nil {
		logger.Fatal("Run aggregate error: ", err)
	}

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
		lsn.Close()
	}()

	return nil
}
