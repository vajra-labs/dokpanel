package health

type MemoryInfo struct {
	Bytes uint64 `json:"bytes"`
	Human string `json:"human"`
}

type MemoryUsage struct {
	Alloc      MemoryInfo `json:"alloc"`
	TotalAlloc MemoryInfo `json:"totalAlloc"`
	Sys        MemoryInfo `json:"sys"`
	HeapAlloc  MemoryInfo `json:"heapAlloc"`
	HeapSys    MemoryInfo `json:"heapSys"`
}

type HealthRes struct {
	Uptime    string      `json:"uptime"`
	Version   string      `json:"version"`
	GoEnv     string      `json:"environment"`
	Timestamp string      `json:"timestamp"`
	Memory    MemoryUsage `json:"memory"`
}
