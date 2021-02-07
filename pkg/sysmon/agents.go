package sysmon

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
)

var cpuStatPrev []uint64 = []uint64{0, 0, 0, 0}

// Function read /proc/loadavg
func getLA() (float32, error) {
	var avg float32 = 0
	loadavg, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	fmt.Sscanf(string(loadavg), "%f", &avg)
	return avg, nil
}

func getCpu() ([3]float32, error) {
	cpuStatLast := make([]uint64, 4)
	cpuStatDelta := make([]uint64, 4)
	cpuStatFields := make([]uint64, 3)
	cpuSumm := uint64(0)
	cpu := [3]float32{0, 0, 0}

	stat, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return cpu, err
	}
	fields := strings.Fields(strings.Split(string(stat), "\n")[0])
	for i, field := range fields[1:5] {
		parsedField, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpu, err
		}
		cpuStatLast[i] = parsedField
	}
	if cpuStatPrev[3]+cpuStatPrev[2] == 0 {
		_ = copy(cpuStatPrev, cpuStatLast)
		return cpu, nil
	}
	for i, _ := range cpuStatLast {
		cpuStatDelta[i] = cpuStatLast[i] - cpuStatPrev[i]
	}
	for _, val := range cpuStatDelta {
		cpuSumm += val
	}
	cpuStatFields[0] = cpuStatDelta[0]
	cpuStatFields[1] = cpuStatDelta[2]
	cpuStatFields[2] = cpuStatDelta[3]
	for i, field := range cpuStatFields {
		cpu[i] = float32(field) * 100 / float32(cpuSumm)
	}
	return cpu, nil
}

func getDiskLoad() ([2]float32, error) {
	diskLoad := [2]float32{0, 0}
	return diskLoad, nil
}

func getFsUsage() (map[string][2]float32, error) {
	fsUsage := map[string][2]float32{"/": {0, 0}}
	return fsUsage, nil
}

/*
   // Type for store OS network listner
   type listner struct {
   	cmd   string
   	pid   uint
   	user  string
   	proto uint
   	port  uint
   }
*/
func getNetListner() (listner, error) {
	netListner := new(listner)
	return *netListner, nil
}

func getNetSocks() (uint, error) {
	var netSocks uint = 0
	return netSocks, nil
}

func runAgents(doneCh <-chan interface{}, cond *sync.Cond, logger *log.Logger) (map[string]chan monState, error) {
	// difine chanels betwen agent and agregator.
	agents := make(map[string]chan monState)
	agents["loadAver"] = make(chan monState)
	agents["cpu"] = make(chan monState)
	agents["diskLoad"] = make(chan monState)
	agents["fsUsage"] = make(chan monState)
	agents["netListner"] = make(chan monState)
	agents["netSocks"] = make(chan monState)
	//difine agents.
	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		for {
			select {
			default:
				cond.L.Lock()
				cond.Wait()
				cond.L.Unlock()
				out, err := getLA()
				if err != nil {
					logger.Printf("ERROR loadAverage agent: %v", err)
				}
				agents["loadAver"] <- monState{loadAver: out}
			case <-doneCh:
				wg.Done()
				return
			}
		}
	}()
	go func() {
		for {
			select {
			default:
				cond.L.Lock()
				cond.Wait()
				cond.L.Unlock()
				out, err := getCpu()
				if err != nil {
					logger.Printf("ERROR cpu agent: %v", err)
				}
				agents["cpu"] <- monState{cpu: out}
			case <-doneCh:
				wg.Done()
				return
			}
		}
	}()
	go func() {
		for {
			select {
			default:
				cond.L.Lock()
				cond.Wait()
				cond.L.Unlock()
				out, err := getDiskLoad()
				if err != nil {
					logger.Printf("ERROR diskLoad agent: %v", err)
				}
				agents["diskLoad"] <- monState{diskLoad: out}
			case <-doneCh:
				wg.Done()
				return
			}
		}
	}()
	go func() {
		for {
			select {
			default:
				cond.L.Lock()
				cond.Wait()
				cond.L.Unlock()
				out, err := getFsUsage()
				if err != nil {
					logger.Printf("ERROR fsUsage agent: %v", err)
				}
				agents["fsUsage"] <- monState{fsUsage: out}
			case <-doneCh:
				wg.Done()
				return
			}
		}
	}()
	go func() {
		for {
			select {
			default:
				cond.L.Lock()
				cond.Wait()
				cond.L.Unlock()
				out, err := getNetListner()
				if err != nil {
					logger.Printf("ERROR netListner agent: %v", err)
				}
				agents["netListner"] <- monState{netListner: out}
			case <-doneCh:
				wg.Done()
				return
			}
		}
	}()
	go func() {
		for {
			select {
			default:
				cond.L.Lock()
				cond.Wait()
				cond.L.Unlock()
				out, err := getNetSocks()
				if err != nil {
					logger.Printf("ERROR netSocks agent: %v", err)
				}
				agents["netSocks"] <- monState{netSocks: out}
			case <-doneCh:
				wg.Done()
				return
			}
		}
	}()
	go func() {
		wg.Wait()
		close(agents["loadAver"])
		close(agents["cpu"])
		close(agents["diskLoad"])
		close(agents["fsUsage"])
		close(agents["netListner"])
		close(agents["netSocks"])
		return
	}()
	return agents, nil
}
