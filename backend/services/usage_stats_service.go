package services

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// UsageStatsService 使用统计服务
type UsageStatsService struct {
	apiURL      string
	version     string
	deviceID    string
	appName     string
	os          string
	arch        string
	hostname    string
	enabled     bool
	stopChan    chan struct{}
	once        sync.Once
	mu          sync.RWMutex
}

// StatsRequest 统计请求结构
type StatsRequest struct {
	AppName   string `json:"app_name,omitempty"`  // 应用名称
	Version   string `json:"version"`
	DeviceID  string `json:"device_id,omitempty"`  // 设备唯一标识码
	OS        string `json:"os,omitempty"`         // 操作系统
	Arch      string `json:"arch,omitempty"`       // 架构
	Hostname  string `json:"hostname,omitempty"`   // 主机名
}

// NewUsageStatsService 创建使用统计服务
func NewUsageStatsService(version string) *UsageStatsService {
	// 获取系统信息
	hostname, _ := os.Hostname()
	
	// 检查是否禁用统计
	enabled := os.Getenv("DISABLE_STATS") != "true"
	
	return &UsageStatsService{
		apiURL:   "", // API地址，需要用户填写
		version:  version,
		os:       runtime.GOOS,
		arch:     runtime.GOARCH,
		hostname: hostname,
		enabled:  enabled,
		stopChan: make(chan struct{}),
	}
}

// SetAppName 设置应用名称
func (s *UsageStatsService) SetAppName(appName string) {
	s.appName = appName
}

// GetDeviceID 获取当前设备标识码
func (s *UsageStatsService) GetDeviceID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deviceID
}

// InitDeviceID 初始化设备标识码（内部获取并存储）
func (s *UsageStatsService) InitDeviceID(dataDir string) error {
	deviceID, err := s.getDeviceID(dataDir)
	if err != nil {
		return err
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceID = deviceID
	return nil
}

// getDeviceID 获取设备唯一标识码
func (s *UsageStatsService) getDeviceID(dataDir string) (string, error) {
	// 优先从文件读取已保存的设备ID
	savedID, err := s.readSavedDeviceID(dataDir)
	if err == nil && savedID != "" {
		return savedID, nil
	}

	// 生成新的设备ID
	newID, err := s.generateDeviceID()
	if err != nil {
		return "", err
	}

	// 保存设备ID
	if err := s.saveDeviceID(dataDir, newID); err != nil {
		// 即使保存失败，也返回生成的ID
		return newID, nil
	}

	return newID, nil
}

// readSavedDeviceID 从文件读取已保存的设备ID
func (s *UsageStatsService) readSavedDeviceID(dataDir string) (string, error) {
	deviceIDFile := filepath.Join(dataDir, "device_id.txt")
	data, err := os.ReadFile(deviceIDFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// saveDeviceID 保存设备ID到文件
func (s *UsageStatsService) saveDeviceID(dataDir, deviceID string) error {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	deviceIDFile := filepath.Join(dataDir, "device_id.txt")
	return os.WriteFile(deviceIDFile, []byte(deviceID), 0644)
}

// generateDeviceID 生成设备唯一标识码
func (s *UsageStatsService) generateDeviceID() (string, error) {
	var systemInfo strings.Builder

	// 1. 操作系统和架构
	fmt.Fprintf(&systemInfo, "OS:%s Arch:%s", runtime.GOOS, runtime.GOARCH)

	// 2. 主机名
	hostname, err := os.Hostname()
	if err == nil {
		fmt.Fprintf(&systemInfo, " Hostname:%s", hostname)
	}

	// 3. 根据操作系统获取更多系统信息
	switch runtime.GOOS {
	case "darwin": // macOS
		if serial, err := s.getMacSerial(); err == nil {
			fmt.Fprintf(&systemInfo, " Serial:%s", serial)
		}
	case "linux":
		if machineID, err := s.getLinuxMachineID(); err == nil {
			fmt.Fprintf(&systemInfo, " MachineID:%s", machineID)
		}
		if productUUID, err := s.getLinuxProductUUID(); err == nil {
			fmt.Fprintf(&systemInfo, " ProductUUID:%s", productUUID)
		}
	case "windows":
		if machineGUID, err := s.getWindowsMachineGUID(); err == nil {
			fmt.Fprintf(&systemInfo, " MachineGUID:%s", machineGUID)
		}
	}

	// 4. 生成 MD5 哈希作为设备ID
	hash := md5.Sum([]byte(systemInfo.String()))
	return hex.EncodeToString(hash[:]), nil
}

// getMacSerial 获取 macOS 序列号
func (s *UsageStatsService) getMacSerial() (string, error) {
	cmd := exec.Command("ioreg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IOPlatformSerialNumber") {
			parts := strings.Split(line, "\"")
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("serial number not found")
}

// getLinuxMachineID 获取 Linux 机器 ID
func (s *UsageStatsService) getLinuxMachineID() (string, error) {
	data, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// getLinuxProductUUID 获取 Linux 产品 UUID
func (s *UsageStatsService) getLinuxProductUUID() (string, error) {
	data, err := os.ReadFile("/sys/class/dmi/id/product_uuid")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// getWindowsMachineGUID 获取 Windows 机器 GUID
func (s *UsageStatsService) getWindowsMachineGUID() (string, error) {
	cmd := exec.Command("wmic", "csproduct", "get", "uuid")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		uuid := strings.TrimSpace(lines[1])
		if uuid != "" {
			return uuid, nil
		}
	}

	return "", fmt.Errorf("machine GUID not found")
}

// SetAPIURL 设置统计API地址
func (s *UsageStatsService) SetAPIURL(url string) {
	s.apiURL = url
}

// Start 启动统计服务
func (s *UsageStatsService) Start() {
	if s.apiURL == "" || !s.enabled {
		return
	}

	// 立即执行一次
	go s.sendStats()

	// 启动定时器，每天不定时执行一次
	go s.dailyScheduler()
}

// Stop 停止统计服务
func (s *UsageStatsService) Stop() {
	s.once.Do(func() {
		close(s.stopChan)
	})
}

// dailyScheduler 每天不定时执行统计
func (s *UsageStatsService) dailyScheduler() {
	// Go 1.20+ 不需要显式设置随机种子，全局随机生成器已自动初始化

	for {
		// 计算到下一次执行的时间：随机选择明天0点-24点之间的某个时刻
		now := time.Now()
		tomorrow := now.AddDate(0, 0, 1)
		midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

		// 随机延迟 0-24 小时
		randomDelay := time.Duration(rand.Intn(24)) * time.Hour
		nextRun := midnight.Add(randomDelay)

		// 使用 NewTimer 而不是 time.After，确保可以及时释放资源
		timer := time.NewTimer(time.Until(nextRun))

		select {
		case <-timer.C:
			// 时间到了，发送统计
			go s.sendStats()
		case <-s.stopChan:
			// 停止信号，退出循环
			timer.Stop() // 停止 timer，防止泄漏
			return
		}
	}
}

// sendStats 发送统计数据
func (s *UsageStatsService) sendStats() {
	// 使用读锁，防止读取时 apiURL 被修改
	s.mu.RLock()
	apiURL := s.apiURL
	version := s.version
	deviceID := s.deviceID
	appName := s.appName
	osType := s.os
	arch := s.arch
	hostname := s.hostname
	s.mu.RUnlock()

	if apiURL == "" {
		return
	}

	req := StatsRequest{
		AppName:  appName,
		Version:  version,
		DeviceID: deviceID,
		OS:       osType,
		Arch:     arch,
		Hostname: hostname,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return // 静默失败
	}

	// 创建 HTTP 请求，设置10秒超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return // 静默失败
	}

	// 确保关闭响应体
	if resp != nil {
		// 读取并丢弃响应体，确保连接可以被复用
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	// 不处理返回结果，不记录日志
}
