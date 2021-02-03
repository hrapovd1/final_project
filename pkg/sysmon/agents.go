package sysmon

/*
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
*/

func runAgents(doneCh <-chan interface{}) (map[string]chan monState, error) {
	agents := make(map[string]chan monState)
	agents["loadAver"] = make(chan monState)
	agents["cpu"] = make(chan monState)
	agents["diskLoad"] = make(chan monState)
	agents["fsUsage"] = make(chan monState)
	agents["netListner"] = make(chan monState)
	agents["netSocks"] = make(chan monState)
	return agents, nil
}
