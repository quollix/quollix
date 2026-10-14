package metrics

const (
	containerMetricPrefix               = "container_"
	ContainerCPUUsageSecondsMetric      = containerMetricPrefix + "cpu_usage_seconds_total"
	ContainerMemoryUsageBytesMetric     = containerMetricPrefix + "memory_usage_bytes"
	ContainerMemoryLimitBytesMetric     = containerMetricPrefix + "memory_limit_bytes"
	ContainerNetworkReceivedBytesMetric = containerMetricPrefix + "network_received_bytes_total"
	ContainerNetworkSentBytesMetric     = containerMetricPrefix + "network_sent_bytes_total"
	ContainerDiskReadBytesMetric        = containerMetricPrefix + "disk_read_bytes_total"
	ContainerDiskWrittenBytesMetric     = containerMetricPrefix + "disk_written_bytes_total"
)

const (
	VolumeUsedBytesMetric = "volume_used_bytes"
)

const (
	MetricNameLabel = "__name__"
	MaintainerLabel = "maintainer"
	AppLabel        = "app"
	ServiceLabel    = "service"
	VolumeLabel     = "volume"
)
