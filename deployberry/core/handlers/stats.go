package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func CheckResourceUsage(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"memory":       usedMem,
		"total_memory": totalMem,
		"cpu":          cpuPercent,
		"disk_usage":   usedDisk,
		"disk_total":   totalDisk,
	})
}
