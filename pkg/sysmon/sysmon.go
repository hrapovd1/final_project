package sysmon

import (
	"log"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	smgrpc "github.com/hrapovd1/final_project/pkg/smgrpc"
	grpc "google.golang.org/grpc"
	metadata "google.golang.org/grpc/metadata"
)

// Internal state of sys-mon process.
type sysmonState struct {
	port       uint
	dataBuff   uint
	answPeriod uint
}

// monitoring server state.
var smState *sysmonState

// Type for store OS network listner.
type listner struct {
	/*
		cmd   string
		pid   uint
		user  string
		proto uint
		port  uint
	*/
}

// Type for store monitoring data scrape.
type monState struct {
	loadAver   float32
	cpu        [3]float32
	diskLoad   map[string][]float32
	fsUsage    map[string][2]float64
	netListner listner
	netSocks   uint
}

// Monitoring internal buffer.
type monStateBuff struct {
	mu   sync.RWMutex
	buff []monState
}

// System monitoring process.
type Sysmon interface {
	Run(doneCh <-chan interface{}, logger *log.Logger)
}

// System monitoring constructor.
func NewSysmon(dataBuff, answPeriod, port uint) Sysmon {
	smState = &sysmonState{
		port:       port,
		dataBuff:   dataBuff,
		answPeriod: answPeriod,
	}
	smBuff := new(monStateBuff)
	smBuff.buff = make([]monState, dataBuff) // sys-mon scrapes buffer

	return smBuff
}

func (mst *monStateBuff) runAgents(doneCh <-chan interface{}, logger *log.Logger) {
	countTmp := 0
	count := &countTmp
	condMu := new(sync.Mutex)
	cond := sync.NewCond(condMu)
	second := time.NewTicker(time.Second)

	go func() { // syncro threat for agents
		for {
			select {
			case <-second.C:
				*count++
				if *count == len(mst.buff) {
					*count = 0
				}
				cond.Broadcast()
			case <-doneCh:
				second.Stop()
				break
			}
		}
	}()

	/* This functions are defined in agents.go
	   because linter error: too many statements
	*/
	go mst.fillLA(doneCh, cond, count, logger)         // get load average threat
	go mst.fillCPU(doneCh, cond, count, logger)        // get cpu statistics threat
	go mst.fillDiskLoad(doneCh, cond, count, logger)   // get disks statistics threat
	go mst.fillFsUsage(doneCh, cond, count, logger)    // get fs usage threat
	go mst.fillNetListner(doneCh, cond, count, logger) // get network listeners threat
	go mst.fillNetSocks(doneCh, cond, count, logger)   // get network sockets data threat
}

// grpc server implementation.
type statServer struct {
	smgrpc.UnimplementedStatServer
	monBuff *monStateBuff
}

func (sS *statServer) GetAll(query *smgrpc.Request, out smgrpc.Stat_GetAllServer) error {
	answer := time.NewTicker(time.Duration(smState.answPeriod) * time.Second)
	header := metadata.New(map[string]string{"application": "System monitoring", "timestamp": time.Now().String()})
	err := out.SendHeader(header)
	if err != nil {
		answer.Stop()
		return err
	}

	for {
		msg := make([]smgrpc.All, len(sS.monBuff.buff))
		for i := range msg {
			// loop for fill All message from monitoring buffer.
			scrape := sS.monBuff.buff[i]
			// aggregate disks data
			disksCount := len(scrape.diskLoad)
			disksLoad := make([]*smgrpc.Disk, disksCount)
			sS.monBuff.mu.RLock()
			for disk, stats := range scrape.diskLoad {
				disksLoad[disksCount-1] = &smgrpc.Disk{Name: disk, Tps: float32(math.Round(float64(stats[0])*100) / 100), Kbps: float32(math.Round(float64(stats[1])*100) / 100)}
				disksCount--
			}
			sS.monBuff.mu.RUnlock()
			// aggregate file systems usage data.
			fssCount := len(scrape.fsUsage)
			fsUsage := make([]*smgrpc.Fs, fssCount)
			sS.monBuff.mu.RLock()
			for fsPath, usage := range scrape.fsUsage {
				fsUsage[fssCount-1] = &smgrpc.Fs{Name: fsPath, Used: float32(math.Round(usage[0]*100) / 100), Iused: float32(math.Round(usage[1]*100) / 100)}
				fssCount--
			}
			sS.monBuff.mu.RUnlock()
			// fill load average.
			sS.monBuff.mu.RLock()
			msg[i].LoadAverage = &smgrpc.LoadAverage{Load: scrape.loadAver}
			sS.monBuff.mu.RUnlock()
			// fill cpu usage.
			sS.monBuff.mu.RLock()
			msg[i].Cpu = &smgrpc.Cpu{Sys: scrape.cpu[0], User: scrape.cpu[1], Idle: scrape.cpu[2]}
			sS.monBuff.mu.RUnlock()
			// fill disk load.
			msg[i].Disk = disksLoad
			// fill network connections.
			msg[i].Connections = &smgrpc.TcpConnections{Count: 10}
			// fill file systems data usage.
			msg[i].Partitions = fsUsage
			// fill network listeners in system.
			msg[i].Listeners = []*smgrpc.Listen{
				{Cmd: "bind", User: "nobody", Pid: 456, Proto: "Tcp", Port: 53},
				{Cmd: "sys-mon", User: "dima", Pid: 7654, Proto: "Tcp", Port: 8080},
			}
			// fill top network protocol talker.
			msg[i].ProtoTalkers = []*smgrpc.NetProtoTalker{
				{Proto: "Tcp", Bytes: 456789, Rate: 100},
				{Proto: "ICMP", Bytes: 12345, Rate: 50},
			}
			// fill top network talker.
			msg[i].RateTalker = []*smgrpc.NetRateTalker{
				{Proto: "Tcp", Sport: 3456, Dport: 80, Bps: 30},
				{Proto: "Tcp", Sport: 4321, Dport: 80, Bps: 28},
			}
		}
		<-answer.C
		for _, message := range msg {
			err := out.Send(&message)
			if err != nil {
				answer.Stop()
				return err
			}
		}
	}
}

// sys-mon implementation.
func (mst *monStateBuff) Run(doneCh <-chan interface{}, logger *log.Logger) {
	mst.runAgents(doneCh, logger)

	srvSock := ":" + strconv.Itoa(int(smState.port))
	lsn, err := net.Listen("tcp", srvSock)
	if err != nil {
		logger.Fatal(err)
	}

	monServer := new(statServer)
	monServer.monBuff = mst

	server := grpc.NewServer()
	smgrpc.RegisterStatServer(server, monServer)

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
	}()
}
