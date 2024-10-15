package main

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"net/http"
)

type SystemInfo struct {
	CPU    []string `json:"cpu"`
	Memory string   `json:"memory"`
	Disk   string   `json:"disk"`
	Uptime string   `json:"uptime"`
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatUptime(seconds uint64) string {
	days := seconds / (24 * 3600)
	hours := (seconds % (24 * 3600)) / 3600
	minutes := (seconds % 3600) / 60
	sec := seconds % 60
	return fmt.Sprintf("%d days, %d hours, %d minutes, %d seconds", days, hours, minutes, sec)
}

func getSystemInfo() (SystemInfo, error) {
	// CPU-Auslastung
	cpuPercents, err := cpu.Percent(0, true) // true für jeden CPU-Kern
	if err != nil {
		return SystemInfo{}, err
	}
	var cpuDetails []string
	for i, percent := range cpuPercents {
		cpuDetails = append(cpuDetails, fmt.Sprintf("CPU %d: %.2f%%", i+1, percent))
	}

	// RAM-Nutzung
	memory, err := mem.VirtualMemory()
	if err != nil {
		return SystemInfo{}, err
	}

	// Festplattenspeicher
	diskUsage, err := disk.Usage("/")
	if err != nil {
		return SystemInfo{}, err
	}

	// System-Uptime
	uptime, err := host.Uptime()
	if err != nil {
		return SystemInfo{}, err
	}

	return SystemInfo{
		CPU:    cpuDetails,
		Memory: formatBytes(memory.Used),
		Disk:   formatBytes(diskUsage.Used),
		Uptime: formatUptime(uptime),
	}, nil
}

func systemInfoHandler(w http.ResponseWriter, r *http.Request) {
	info, err := getSystemInfo()
	if err != nil {
		http.Error(w, "Unable to get system info", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	http.HandleFunc("/info", systemInfoHandler)
	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
