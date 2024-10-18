package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL        = "http://srv.msk01.gigacorp.local/_stats"
	loadAvgThreshold = 30
	memoryUsageLimit = 1  // 80%
	diskUsageLimit   = 90 // 90%
	netUsageLimit    = 90 // 90%
)

func fetchServerStats() (string, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("Error closing response body: %v\n", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseStats(data string) ([]int, error) {
	parts := strings.Split(data, ",")
	stats := make([]int, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part) // Убираем возможные пробелы
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		stats[i] = value
	}

	return stats, nil
}

func checkThresholds(stats []int) {
	loadAvg := stats[0]
	totalMemory := stats[1]
	usedMemory := stats[2]
	totalDisk := stats[3]
	usedDisk := stats[4]
	totalNet := stats[5]
	usedNet := stats[6]

	// Load Average check
	if loadAvg > loadAvgThreshold {
		fmt.Printf("Load Average is too high: %d\n", loadAvg)
	}

	// Memory usage check
	memoryUsage := (usedMemory / totalMemory) * 100
	if memoryUsage > memoryUsageLimit {
		fmt.Printf("Memory usage too high: %d%%\n", memoryUsage)
	}

	// Disk space check
	freeDiskMb := (totalDisk - usedDisk) / 1000000
	if (usedDisk/totalDisk)*100 > diskUsageLimit {
		fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskMb)
	}

	// Network bandwidth check
	freeNet := (totalNet - usedNet) / 1000000
	if (usedNet/totalNet)*100 > netUsageLimit {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeNet)
	}
}

func main() {
	errorCount := 0
	for {
		statsData, err := fetchServerStats()
		if err != nil {
			fmt.Println("Error fetching server statistics:", err)
			errorCount++
		} else {
			stats, err := parseStats(statsData)
			if err != nil {
				fmt.Println("Error parsing server statistics:", err)
				errorCount++
			} else {
				checkThresholds(stats)
				errorCount = 0
			}
		}

		if errorCount >= 3 {
			fmt.Println("Unable to fetch server statistics")
		}

		// Sleep for 10 seconds before the next request
		time.Sleep(10 * time.Second)
	}
}
