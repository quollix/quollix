package metrics

import "time"

type ResourceUsage struct {
	CPUUsageSeconds      float64 `json:"cpu_usage_seconds_total"`
	MemoryUsageBytes     uint64  `json:"mem_bytes"`
	MemoryLimitBytes     uint64  `json:"mem_limit_bytes"`
	NetworkReceivedBytes uint64  `json:"net_rx_bytes"`
	NetworkSentBytes     uint64  `json:"net_tx_bytes"`
	DiskReadBytes        uint64  `json:"blk_read_bytes"`
	DiskWrittenBytes     uint64  `json:"blk_write_bytes"`
}

type ContainerMetricsSample struct {
	Timestamp  time.Time     `json:"timestamp"`
	Maintainer string        `json:"maintainer"`
	App        string        `json:"app"`
	Service    string        `json:"service"`
	Usage      ResourceUsage `json:"data"`
}
