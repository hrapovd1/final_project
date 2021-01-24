package sysmon

import (
	"log"
	"os"
	"time"
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

func NewSysmon(dataBuff, answPeriod, port *uint) Sysmon {
	smState = &sysmonState{
		port:       &port,
		dataBuff:   &dataBuff,
		answPeriod: &answPeriod,
	}
	return &monState{
		loadAver:   0,
		cpu:        {0, 0, 0},
		diskLoad:   {0, 0},
		fsUsage:    make(map[string][2]uint, 1),
		netListner: make(listner),
		netSocks:   0,
	}
}

func (mst *monState) Run(logger log.Logger, sysCh <-chan os.Signal) error {
	go func(sysCh <-chan os.Signal) {
		select {
		case <-sysCh:
			return
		default:
			for {
				time.Sleep(time.Duration(*smState.dataBuff) * time.Second)
				logger.Println("run load data getter")
			}
		}
	}(sysCh)
	go func(sysCh <-chan os.Signal) {
		select {
		case <-sysCh:
			return
		default:
			for {
				time.Sleep(time.Duration(*smState.dataBuff) * time.Second)
				logger.Println("run cpu data getter")
			}
		}
	}(sysCh)
	go func(sysCh <-chan os.Signal) {
		select {
		case <-sysCh:
			return
		default:
			for {
				time.Sleep(time.Duration(*smState.dataBuff) * time.Second)
				logger.Println("run disk data getter")
			}
		}
	}(sysCh)
	go func(sysCh <-chan os.Signal) {
		select {
		case <-sysCh:
			return
		default:
			for {
				time.Sleep(time.Duration(*smState.dataBuff) * time.Second)
				logger.Println("run fs data getter")
			}
		}
	}(sysCh)
	go func(sysCh <-chan os.Signal) {
		select {
		case <-sysCh:
			return
		default:
			for {
				time.Sleep(time.Duration(*smState.dataBuff) * time.Second)
				logger.Println("run netstat data getter")
			}
		}
	}(sysCh)
	go func(sysCh <-chan os.Signal) {
		select {
		case <-sysCh:
			return
		default:
			for {
				time.Sleep(time.Duration(*smState.dataBuff) * time.Second)
				logger.Println("run data aggregator")
			}
		}
	}(sysCh)
	go func(sysCh <-chan os.Signal) {
		select {
		case <-sysCh:
			return
		default:
			for {
				time.Sleep(time.Duration(*smState.answPeriod) * time.Second)
				logger.Printf("open %v port to listen...", &smState.port)
			}
		}
	}(sysCh)

	return nil
}
