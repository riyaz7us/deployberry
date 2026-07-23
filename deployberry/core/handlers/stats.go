package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func CheckResourceUsage(c *gin.Context) {
	// Set mandatory headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	isFirst := true
	c.Stream(func(w io.Writer) bool {
		if !isFirst {
			time.Sleep(7 * time.Second)
		} else {
			isFirst = false
		}
		stats, _ := checkResourceUsage()

		jsonData, err := json.Marshal(stats)
		if err != nil {
			return false // Stop streaming if marshalling fails
		}

		// SSE requires the format: "data: <your-message>\n\n"
		message := fmt.Sprintf("data: %v\n\n", string(jsonData))

		fmt.Println("[SSE] Sending stats event...")
		if _, err := w.Write([]byte(message)); err != nil {
			fmt.Printf("[SSE] Connection closed/broken: %v\n", err)
			return false // Stop streaming if connection is broken
		}
		return true // Continue streaming
	})
}

func checkResourceUsage() (map[string]interface{}, error) {
	cpuPercent, _ := cpu.Percent(0, true)

	var usedMem float64
	var totalMem uint64
	if memStats, err := mem.VirtualMemory(); err == nil && memStats != nil {
		usedMem = memStats.UsedPercent
		totalMem = memStats.Total / 1024 / 1024 // MB
	}

	var usedDisk float64
	var totalDisk uint64
	if diskStats, err := disk.Usage("/"); err == nil && diskStats != nil {
		usedDisk = diskStats.UsedPercent
		totalDisk = diskStats.Total / 1024 / 1024 / 1024 // GB
	}
	return map[string]interface{}{
		"memory":       usedMem,
		"total_memory": totalMem,
		"cpu":          cpuPercent,
		"disk_usage":   usedDisk,
		"disk_total":   totalDisk,
	}, nil
}
