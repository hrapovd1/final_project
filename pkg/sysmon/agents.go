package sysmon

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var cpuStatPrev []uint64 = make([]uint64, 4)
var diskStatPrev map[string][]uint64 = make(map[string][]uint64, 0)
var uptimePrev float64 = float64(0)

// Function read /proc/loadavg
func getLA() (float32, error) {
	var avg float32 = 0
	loadavg, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return avg, err
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

// tps (transfers per second); KB/s (kilobytes (read+write) per second);
func getDiskLoad() (map[string][]float32, error) {
	diskLoad := make(map[string][]float32, 0)
	diskStatLast := make(map[string][]uint64, 0)
	uptimeLast := float64(0)

	// find block devices
	files, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		return diskLoad, err
	}
	// filter loops
	for _, item := range files {
		if !strings.Contains(item.Name(), "loop") {
			diskStatLast[item.Name()] = make([]uint64, 16)
		}
	}
	// count measure interval acording uptime
	// This interval count method is got from
	// https://github.com/sysstat/sysstat/blob/master/iostat.c
	uptime, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return diskLoad, err
	}
	valuesStr := strings.Split(strings.Trim(string(uptime), "\n"), " ")
	valuesFloat1, err := strconv.ParseFloat(valuesStr[0], 64)
	if err != nil {
		return diskLoad, err
	}
	valuesFloat2, err := strconv.ParseFloat(valuesStr[1], 64)
	if err != nil {
		return diskLoad, err
	}
	uptimeLast = valuesFloat1 + valuesFloat2/100 //this is coefficient from iostat
	if uptimePrev == 0 {
		uptimePrev = uptimeLast - 1
	}
	itv := uptimeLast - uptimePrev
	uptimePrev = uptimeLast

	// get statistics for block devices
	// This statistics count method is got from
	// https://github.com/sysstat/sysstat/blob/master/iostat.c
	diskStats, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return diskLoad, err
	}
	/*
		https://gist.github.com/lesovsky/e150e82d97ad691dbbfd
		https://github.com/sysstat/sysstat/blob/master/iostat.c
			   S_VALUE(ioj->rd_ios + ioj->wr_ios + ioj->dc_ios,
			   ioi->rd_ios + ioi->wr_ios + ioi->dc_ios, itv));
	*/
	for _, line := range strings.Split(string(diskStats), "\n") {
		if len(line) == 0 {
			continue
		}
		fields := strings.Fields(line)
		if !strings.Contains(fields[2], "loop") {
			for disk, _ := range diskStatLast {
				if strings.EqualFold(fields[2], disk) {
					for i, field := range fields[3:] {
						fieldUint, err := strconv.ParseUint(field, 10, 64)
						if err != nil {
							return diskLoad, err
						}
						diskStatLast[disk][i] = fieldUint
					}
				}
			}
		}
	}
	if len(diskStatPrev) == 0 {
		diskStatPrev = diskStatLast
		return diskLoad, nil
	}
	for disk, _ := range diskStatLast {
		deltaTps := float64((diskStatLast[disk][0] + diskStatLast[disk][4] + diskStatLast[disk][8]) -
			(diskStatPrev[disk][0] + diskStatPrev[disk][4] + diskStatPrev[disk][8]))
		tps := deltaTps / itv
		deltaKbs := float64((diskStatLast[disk][2] + diskStatLast[disk][6] + diskStatLast[disk][10]) -
			(diskStatPrev[disk][2] + diskStatPrev[disk][6] + diskStatPrev[disk][10]))
		kbs := deltaKbs / itv / 2
		diskLoad[disk] = []float32{float32(tps), float32(kbs)}
	}
	diskStatPrev = diskStatLast
	return diskLoad, nil
}

func getFsUsage() (map[string][2]float64, error) {
	fsUsage := map[string][2]float64{"": {0, 0}}
	fsPaths := make([]string, 0)
	fsFinder, err := regexp.Compile(`^/[^\s].*`)
	if err != nil {
		return fsUsage, err
	}

	// get mounted fss
	lines, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return fsUsage, err
	}
	// filter loops
	for _, line := range strings.Split(string(lines), "\n") {
		if len(line) == 0 || strings.Contains(line, "loop") {
			continue
		}
		if fsFinder.MatchString(line) {
			fields := strings.Fields(line)
			fsPaths = append(fsPaths, fields[1])
		}
	}

	fsUsage = make(map[string][2]float64, len(fsPaths))

	// This count way is brought from
	// https://github.com/google/cadvisor/blob/57a2c804a08755a29e44afa26b4b8e60add4e420/fs/fs.go#L647
	for _, path := range fsPaths {
		var s syscall.Statfs_t
		var usage float64
		var iusage float64
		err := syscall.Statfs(path, &s)
		if err != nil {
			return fsUsage, err
		}

		total := float64(uint64(s.Frsize) * s.Blocks)
		avail := float64(uint64(s.Frsize) * s.Bavail)
		inodes := float64(s.Files)
		inodesFree := float64(s.Ffree)

		if total > 0 {
			usage = (1 - avail/total) * 100
		} else {
			usage = 0
		}
		if inodes > 0 {
			iusage = (1 - inodesFree/inodes) * 100
		} else {
			iusage = 0
		}
		fsUsage[path] = [...]float64{usage, iusage}
	}
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

func runAgents(doneCh <-chan interface{}, cond *sync.Cond, logger *log.Logger) map[string]chan monState {
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
	return agents
}
