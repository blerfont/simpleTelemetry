package main

import (
	"bufio"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	sigar "github.com/cloudfoundry/gosigar"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	loadHelper, _ := load.Avg()
	uptime, _ := host.Uptime()
	memoryHelper, memoryErr := memory.Get()
	if memoryErr != nil {
		return
	}
	totalSpace, usedSpace := getDataSpace("/")


	telemetryJSON := gabs.New()
	_, _ = telemetryJSON.Set(time.Now().UTC().Unix(), "utc")
	_, _ = telemetryJSON.Set(loadHelper.Load5, "cpuUsage")
	_, _ = telemetryJSON.Set(strconv.Itoa(int(memoryHelper.Used/1024/1024))+"MB", "ramUsage")
	_, _ = telemetryJSON.Set(strconv.Itoa(int(getTotalMem()))+"MB", "ramTotal")
	_, _ = telemetryJSON.Set(getCPUTemp(), "cpuTemperature")
	_, _ = telemetryJSON.Set(strconv.Itoa(int(usedSpace))+"GB", "usedDiskSpace")
	_, _ = telemetryJSON.Set(strconv.Itoa(int(totalSpace))+"GB", "totalDiskSpace")
	_, _ = telemetryJSON.Set(time.Now().UTC().Unix()-int64(uptime), "upSince")

	fmt.Println(telemetryJSON)
}

func getCPUTemp() int {
	thermalFile, thermalFileErr := os.Open("/sys/class/thermal/thermal_zone0/temp")
	if thermalFileErr != nil {
		return 0
	}

	defer func() {
		thermalFileCloseErr := thermalFile.Close()
		if thermalFileCloseErr != nil {
			fmt.Println("getCPUTemp - failed to close thermal file: "+thermalFileCloseErr.Error())
		}
	}()

	reader := bufio.NewReader(thermalFile)
	temp, _ := reader.ReadString('\n')
	tempVal, _ := strconv.Atoi(strings.Replace(temp, "\n", "", 1))

	return tempVal / 1000
}

func getTotalMem() uint64 {
	mem := sigar.Mem{}

	memErr := mem.Get()
	if memErr != nil {
		fmt.Println(memErr)
		return 0
	}

	return mem.Total / 1024 / 1024
}

func getDataSpace(partitionMounted string) (uint64, uint64) {
	fs := syscall.Statfs_t{}
	statsFSErr := syscall.Statfs(partitionMounted, &fs)
	if statsFSErr != nil {
		return 0, 0
	}
	totalDataSpace := fs.Blocks * uint64(fs.Bsize) / 1024 / 1024 / 1024
	freeDataSpace := fs.Bfree * uint64(fs.Bsize) / 1024 / 1024 / 1024
	return totalDataSpace, totalDataSpace - freeDataSpace
}