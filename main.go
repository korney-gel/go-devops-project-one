package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	serverURL        = "https://srv.msk01.gigacorp.local/_stats"
	loadAvgThreshold = 30.0
	memoryUsageLimit = 0.80 // 80%
	diskUsageLimit   = 0.90 // 90%
	netUsageLimit    = 0.90 // 90%
)

func fetchServerStats() (string, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseStats(data string) ([]float64, error) {
	parts := strings.Split(data, ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("unexpected data format")
	}

	stats := make([]float64, 7)
	for i, part := range parts {
		value, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, err
		}
		stats[i] = value
	}

	return stats, nil
}

func checkThresholds(stats []float64) {
	loadAvg := stats[0]
	totalMemory := stats[1]
	usedMemory := stats[2]
	totalDisk := stats[3]
	usedDisk := stats[4]
	totalNet := stats[5]
	usedNet := stats[6]

	// Load Average check
	if loadAvg > loadAvgThreshold {
		fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
	}

	// Memory usage check
	memoryUsage := usedMemory / totalMemory
	if memoryUsage > memoryUsageLimit {
		fmt.Printf("Memory usage too high: %.2f%%\n", memoryUsage*100)
	}

	// Disk space check
	freeDisk := (totalDisk - usedDisk) / (1024 * 1024)
	if usedDisk/totalDisk > diskUsageLimit {
		fmt.Printf("Free disk space is too low: %.2f MB left\n", freeDisk)
	}

	// Network bandwidth check
	freeNet := (totalNet - usedNet) / (1024 * 1024) * 8
	if usedNet/totalNet > netUsageLimit {
		fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeNet)
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
		//time.Sleep(10 * time.Second)
	}
}
