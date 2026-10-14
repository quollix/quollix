package metrics

import "time"

type VolumeMetricsSample struct {
	Timestamp  time.Time `json:"timestamp"`
	Maintainer string    `json:"maintainer"`
	App        string    `json:"app"`
	Volume     string    `json:"volume"`
	UsedBytes  uint64    `json:"used_bytes"`
}
