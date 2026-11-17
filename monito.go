package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type Config struct {
	Threshold       float64
	CheckInterval   time.Duration
	CaptureDuration time.Duration
	CaptureDir      string
}

type Snapshot struct {
	Timestamp   string                 `json:"timestamp"`
	Hostname    string                 `json:"hostname"`
	CPU         float64                 `json:"cpu_percent"`
	Memory      *mem.VirtualMemoryStat `json:"memory"`
	Disk        *disk.UsageStat        `json:"disk"`
	Network     []net.IOCountersStat   `json:"network"`
	Processes   []ProcessInfo          `json:"top_processes"`
	SystemInfo  *host.InfoStat         `json:"system_info"`
	Uptime      uint64                  `json:"uptime_seconds"`
}

type ProcessInfo struct {
	PID     int32   `json:"pid"`
	Name    string  `json:"name"`
	CPU     float64 `json:"cpu_percent"`
	Memory  float32 `json:"memory_percent"`
	Cmdline string  `json:"cmdline,omitempty"`
}

type CPUMonitor struct {
	config      Config
	spikeActive bool
	signalChan  chan os.Signal
	doneChan    chan bool
}

func NewCPUMonitor(config Config) *CPUMonitor {
	return &CPUMonitor{
		config:     config,
		signalChan: make(chan os.Signal, 1),
		doneChan:   make(chan bool, 1),
	}
}

func (m *CPUMonitor) Start() {
	log.Printf("🚀 CPU Monitor Started")
	log.Printf("📊 Threshold: %.1f%%", m.config.Threshold)
	log.Printf("⏱️  Check interval: %v", m.config.CheckInterval)
	log.Printf("📁 Capture directory: %s", m.config.CaptureDir)

	// Create capture directory
	os.MkdirAll(m.config.CaptureDir, 0755)

	// Handle shutdown signals
	signal.Notify(m.signalChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkCPU()
		case <-m.signalChan:
			log.Println("👋 Shutting down...")
			m.doneChan <- true
			return
		}
	}
}

func (m *CPUMonitor) checkCPU() {
	percent, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("Error getting CPU: %v", err)
		return
	}

	currentCPU := percent[0]

	if currentCPU > m.config.Threshold && !m.spikeActive {
		m.spikeActive = true
		log.Printf("🔥 SPIKE DETECTED: %.1f%%", currentCPU)
		m.captureSnapshot(currentCPU)
	} else if currentCPU <= m.config.Threshold && m.spikeActive {
		m.spikeActive = false
		log.Printf("✅ CPU returned to normal: %.1f%%", currentCPU)
	}
}

func (m *CPUMonitor) captureSnapshot(cpuPercent float64) {
	timestamp := time.Now().Format("20060102_150405")
	snapDir := filepath.Join(m.config.CaptureDir, fmt.Sprintf("spike_%s", timestamp))
	
	if err := os.MkdirAll(snapDir, 0755); err != nil {
		log.Printf("Failed to create capture dir: %v", err)
		return
	}

	// Collect all data
	snapshot := m.collectData(cpuPercent)

	// Save as JSON
	jsonFile := filepath.Join(snapDir, "snapshot.json")
	data, _ := json.MarshalIndent(snapshot, "", "  ")
	if err := os.WriteFile(jsonFile, data, 0644); err != nil {
		log.Printf("Failed to save JSON: %v", err)
	}

	// Save human-readable process list
	m.saveProcessList(snapDir)

	// Save system logs (Linux only)
	m.saveSystemLogs(snapDir)

	log.Printf("💾 Capture saved to: %s", snapDir)
}

func (m *CPUMonitor) collectData(cpuPercent float64) Snapshot {
	// Memory
	memStat, _ := mem.VirtualMemory()
	
	// Disk
	diskStat, _ := disk.Usage("/")
	
	// Network
	netStats, _ := net.IOCounters(false)
	
	// System info
	hostInfo, _ := host.Info()
	
	// Top processes
	processes := m.getTopProcesses(20)

	return Snapshot{
		Timestamp:   time.Now().Format(time.RFC3339),
		Hostname:    hostInfo.Hostname,
		CPU:         cpuPercent,
		Memory:      memStat,
		Disk:        diskStat,
		Network:     netStats,
		Processes:   processes,
		SystemInfo:  hostInfo,
		Uptime:      hostInfo.Uptime,
	}
}

func (m *CPUMonitor) getTopProcesses(limit int) []ProcessInfo {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	var procList []ProcessInfo
	for _, p := range processes {
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		cmdline, _ := p.Cmdline()

		procList = append(procList, ProcessInfo{
			PID:     p.Pid,
			Name:    name,
			CPU:     cpuPercent,
			Memory:  memPercent,
			Cmdline: cmdline,
		})
	}

	// Sort by CPU descending
	for i := 0; i < len(procList)-1; i++ {
		for j := i + 1; j < len(procList); j++ {
			if procList[i].CPU < procList[j].CPU {
				procList[i], procList[j] = procList[j], procList[i]
			}
		}
	}

	if len(procList) > limit {
		procList = procList[:limit]
	}

	return procList
}

func (m *CPUMonitor) saveProcessList(dir string) {
	file, err := os.Create(filepath.Join(dir, "top_processes.txt"))
	if err != nil {
		return
	}
	defer file.Close()

	fmt.Fprintln(file, "TOP PROCESSES BY CPU USAGE")
	fmt.Fprintln(file, "==========================")
	
	processes := m.getTopProcesses(30)
	for i, p := range processes {
		fmt.Fprintf(file, "%2d. PID:%-6d CPU:%6.2f%% MEM:%5.1f%%  %s\n", 
			i+1, p.PID, p.CPU, p.Memory, p.Name)
	}
}

func (m *CPUMonitor) saveSystemLogs(dir string) {
	// Try to get last 100 lines from syslog (Linux)
	logFile := "/var/log/syslog"
	if _, err := os.Stat(logFile); err == nil {
		data, _ := os.ReadFile(logFile)
		lines := string(data)
		if len(lines) > 10000 {
			lines = lines[len(lines)-10000:]
		}
		os.WriteFile(filepath.Join(dir, "syslog.tail"), []byte(lines), 0644)
	}
}

func main() {
	config := Config{
		Threshold:       85.0,
		CheckInterval:   2 * time.Second,
		CaptureDuration: 30 * time.Second,
		CaptureDir:      "spike_captures",
	}

	// Override with command line flags if needed
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Edit config in code or use environment variables\n")
	}

	monitor := NewCPUMonitor(config)
	monitor.Start()
}