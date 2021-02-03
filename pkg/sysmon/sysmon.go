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
	loadAver   uint
	cpu        [3]uint
	diskLoad   [2]uint
	fsUsage    map[string][2]uint
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
	// sys-mon scrapes buffer
	smBuff := make(monStateBuff, dataBuff)
	return &smBuff
}

// aggregate data from agents and sent "All" messages in grpc
func (mst *monStateBuff) runAggregate(doneCh <-chan interface{}, agents map[string]chan monState, messages chan *smgrpc.All) error {
	count := 0
	answer := time.NewTicker(time.Duration(smState.answPeriod) * time.Second)
	second := time.NewTicker(time.Second)

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
				/*
					type monState struct {
						loadAver   uint
						cpu        [3]uint
						diskLoad   [2]uint
						fsUsage    map[string][2]uint
						netListner listner
						netSocks   uint
					}
				*/
				go func() {
					defer wg.Done()
					agents["loadAver"] <- monState{}
					msg := <-agents["loadAver"]
					(*mst)[count].loadAver = msg.loadAver
				}()
				go func() {
					defer wg.Done()
					agents["cpu"] <- monState{}
					msg := <-agents["loadAver"]
					(*mst)[count].cpu = msg.cpu
				}()
				go func() {
					defer wg.Done()
					agents["diskLoad"] <- monState{}
					msg := <-agents["loadAver"]
					(*mst)[count].diskLoad = msg.diskLoad
				}()
				go func() {
					defer wg.Done()
					agents["fsUsage"] <- monState{}
					msg := <-agents["loadAver"]
					(*mst)[count].fsUsage = msg.fsUsage
				}()
				go func() {
					defer wg.Done()
					agents["netListner"] <- monState{}
					msg := <-agents["loadAver"]
					(*mst)[count].netListner = msg.netListner
				}()
				go func() {
					defer wg.Done()
					agents["netSocks"] <- monState{}
					msg := <-agents["loadAver"]
					(*mst)[count].netSocks = msg.netSocks
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
					   	Disk         *Disk             `protobuf:"bytes,3,opt,name=disk,proto3" json:"disk,omitempty"`
					   	Partitions   []*Fs             `protobuf:"bytes,4,rep,name=partitions,proto3" json:"partitions,omitempty"`
					   	Connections  *TcpConnections   `protobuf:"bytes,5,opt,name=connections,proto3" json:"connections,omitempty"`
					   	Listners     []*Listen         `protobuf:"bytes,6,rep,name=listners,proto3" json:"listners,omitempty"`
					   	ProtoTalkers []*NetProtoTalker `protobuf:"bytes,7,rep,name=protoTalkers,proto3" json:"protoTalkers,omitempty"`
					   	RateTalker   []*NetRateTalker  `protobuf:"bytes,8,rep,name=rateTalker,proto3" json:"rateTalker,omitempty"`
					   }
					*/
					messages <- &smgrpc.All{
						LoadAverage: &smgrpc.LoadAverage{Load: uint32(scrape.loadAver)},
						Cpu:         &smgrpc.Cpu{Sys: uint32(scrape.cpu[0]), User: uint32(scrape.cpu[1]), Idle: uint32(scrape.cpu[2])},
						Disk:        &smgrpc.Disk{Tps: uint32(scrape.diskLoad[0]), Kbps: uint32(scrape.diskLoad[1])},
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
	agentsMap, err := runAgents(doneCh)
	if err != nil {
		logger.Fatal("Run agents error: ", err)
	}

	err = mst.runAggregate(doneCh, agentsMap, allCh)
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
