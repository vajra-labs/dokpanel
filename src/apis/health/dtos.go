package health

// MemoryInfo holds raw bytes and human-readable memory size.
type MemoryInfo struct {
	Bytes uint64 `json:"bytes" doc:"Raw bytes"`
	Human string `json:"human" doc:"Human readable size (e.g. 12 MB)"`
}

// MemoryUsage holds Go runtime memory statistics.
type MemoryUsage struct {
	Alloc      MemoryInfo `json:"alloc"      doc:"Currently allocated heap memory"`
	TotalAlloc MemoryInfo `json:"totalAlloc" doc:"Total allocated heap memory (cumulative)"`
	Sys        MemoryInfo `json:"sys"        doc:"Total memory obtained from OS"`
	HeapAlloc  MemoryInfo `json:"heapAlloc"  doc:"Currently allocated heap objects"`
	HeapSys    MemoryInfo `json:"heapSys"    doc:"Heap memory obtained from OS"`
}

// HealthRes is the response body for the health check endpoint.
type HealthRes struct {
	Uptime    string      `json:"uptime"      doc:"Server uptime (e.g. 2 hours ago)"`
	Version   string      `json:"version"     doc:"Application version"`
	GoEnv     string      `json:"environment" doc:"Current environment (dev, prod, test)"`
	Timestamp string      `json:"timestamp"   doc:"Current ISO 8601 timestamp"`
	Memory    MemoryUsage `json:"memory"      doc:"Go runtime memory usage statistics"`
}
