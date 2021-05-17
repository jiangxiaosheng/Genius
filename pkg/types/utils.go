package types

import (
	"k8s.io/klog/v2"
	"regexp"
	"strconv"
	"strings"
)

var (
	valueRegex      = regexp.MustCompile(`.+=>\s(\d+.?\d*).*`)
	nodeNameRegex   = regexp.MustCompile(`kubernetes_node="([^"]*)"`)
	metricTypeRegex = regexp.MustCompile(`observerward_(\w+)\{`)
	uuidRegex       = regexp.MustCompile(`uuid="([^"]+)"`)
	idRegex         = regexp.MustCompile(`gpu="(\d*)"`)
	modelRegex      = regexp.MustCompile(`model="([^"]+)"`)
)

const (
	decoderUtilizationStr  = "dynamic_gpu_decoder_utilization"
	encoderUtilizationStr  = "dynamic_gpu_encoder_utilization"
	memoryUtilizationStr   = "dynamic_gpu_memory_utilization"
	powerUsageStr          = "dynamic_gpu_power_usage_W"
	usedGlobalMemoryStr    = "dynamic_gpu_used_global_memory_MiB"
	freeGlobalMemoryStr    = "dynamic_gpu_free_global_memory_MiB"
	memorySizeStr          = "static_gpu_memory_size"
	multiprocessorCountStr = "static_gpu_multiprocessor_count"
	sharedDecoderCountStr  = "static_gpu_shared_decoder_count"
	sharedEncoderCountStr  = "static_gpu_shared_encoder_count"
)

var (
	metricTypeMap = map[string]MetricType{
		decoderUtilizationStr:  GPUDecoderUtilization,
		encoderUtilizationStr:  GPUEncoderUtilization,
		memoryUtilizationStr:   GPUMemoryUtilization,
		powerUsageStr:          GPUPowerUsage,
		usedGlobalMemoryStr:    GPUUsedGlobalMemory,
		freeGlobalMemoryStr:    GPUFreeGlobalMemory,
		memorySizeStr:          GPUMemorySize,
		multiprocessorCountStr: GPUMultiprocessorCount,
		sharedDecoderCountStr:  GPUSharedDecoderCount,
		sharedEncoderCountStr:  GPUSharedEncoderCount,
	}
)

func ExtractMetricTypeFromProm(val string) MetricType {
	match := metricTypeRegex.FindStringSubmatch(val)
	if len(match) != 2 {
		klog.Errorf("Extracting Metric Type From Prometheus Query Result Error")
		return -1
	}
	return metricTypeMap[match[1]]
}

func ExtractValueFromProm(val string) uint64 {
	match := valueRegex.FindStringSubmatch(val)
	if len(match) != 2 {
		klog.Errorf("Extracting Value From Prometheus Query Result Error")
		return 0
	}
	strVal := strings.Trim(match[1], " ")
	intVal, _ := strconv.ParseUint(strVal, 10, 64)
	return intVal
}

func ExtractNodeNameFromProm(val string) string {
	match := nodeNameRegex.FindStringSubmatch(val)
	if len(match) != 2 {
		klog.Errorf("Extracting NodeName From Prometheus Query Result Error")
		return ""
	}
	return match[1]
}

func ExtractUUIDFromProm(val string) string {
	match := uuidRegex.FindStringSubmatch(val)
	if len(match) != 2 {
		klog.Errorf("extracting uuid from prometheus query result error")
		return ""
	}
	return match[1]
}

func ExtractIDFromProm(val string) uint {
	match := idRegex.FindStringSubmatch(val)
	if len(match) != 2 {
		klog.Errorf("extracting nodename from prometheus query error")
		return 0
	}
	res, _ := strconv.Atoi(match[1])
	return uint(res)
}

func ExtractModelFromProm(val string) string {
	match := modelRegex.FindStringSubmatch(val)
	if len(match) != 2 {
		klog.Errorf("extracting model from prometheus query error")
		return ""
	}
	return match[1]
}
