package types

import (
	"testing"
)

func TestExtractValueFromProm(t *testing.T) {
	raw := "observerward_gpu_decoder_utilization{UUID=\"GPU-8763e6c0-e8b9-ac77-91ba-407ee16f5493\", gpu=\"1\", hostname=\"observerward-4k68z\", instance=\"192.168.205.76:9909\", job=\"gpu-metrics\", kubernetes_node=\"linww-poweredge-t630\"} => 123 @[1620363129.316]"
	println(ExtractValueFromProm(raw))
}

func TestExtractNodeNameFromProm(t *testing.T) {
	raw := "observerward_gpu_decoder_utilization{UUID=\"GPU-8763e6c0-e8b9-ac77-91ba-407ee16f5493\", gpu=\"1\", hostname=\"observerward-4k68z\", instance=\"192.168.205.76:9909\", job=\"gpu-metrics\", kubernetes_node=\"linww-poweredge-t630\"} => 123 @[1620363129.316]"
	println(ExtractNodeNameFromProm(raw))
}

func TestExtractIDFromProm(t *testing.T) {
	raw := `observerward_gpu_used_global_memory_MiB{UUID="GPU-8763e6c0-e8b9-ac77-91ba-407ee16f5493", gpu="1", hostname="observerward-4k68z", instance="192.168.205.76:9909", job="gpu-metrics", kubernetes_node="linww-poweredge-t630"} => 3740 @[1620548893.055]`
	println(ExtractIDFromProm(raw))
}

func TestExtractUUIDFromProm(t *testing.T) {
	raw := `observerward_gpu_used_global_memory_MiB{UUID="GPU-8763e6c0-e8b9-ac77-91ba-407ee16f5493", gpu="1", hostname="observerward-4k68z", instance="192.168.205.76:9909", job="gpu-metrics", kubernetes_node="linww-poweredge-t630"} => 3740 @[1620548893.055]`
	println(ExtractUUIDFromProm(raw))
}

func TestExtractMetricTypeFromProm(t *testing.T) {
	raw := `observerward_gpu_power_usage_W{UUID="GPU-8763e6c0-e8b9-ac77-91ba-407ee16f5493", gpu="1", hostname="observerward-4k68z", instance="192.168.205.76:9909", job="gpu-metrics", kubernetes_node="linww-poweredge-t630"} => 3740 @[1620548893.055]`
	println(ExtractMetricTypeFromProm(raw))
}

func TestExtractModelFromProm(t *testing.T) {
	raw := `observerward_dynamic_gpu_decoder_utilization{id="0", instance="192.168.205.114:9909", job="gpu-metrics", kubernetes_node="linww-poweredge-t630", model="GeForce GTX 1080 Ti", uuid="GPU-49764fc0-5afa-9237-a573-d226351369f9"} => 0 @[1621255638.419]`
	println(ExtractModelFromProm(raw))
}
