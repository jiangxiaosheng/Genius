package types

import (
	"github.com/observerward/pkg/scraper"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type MetricType int

const (
	_ MetricType = iota
	GPUDecoderUtilization
	GPUEncoderUtilization
	GPUMemoryUtilization
	GPUPowerUsage
	GPUUsedGlobalMemory
	GPUFreeGlobalMemory
	GPUMemorySize
	GPUMultiprocessorCount
	GPUSharedDecoderCount
	GPUSharedEncoderCount
)

const (
	// MetricsTypesCount must match all constant variables of the type MetricType.
	MetricsTypesCount = 10
)

// GPUMetricsWithProm
// key: nodename
// value: metrics of GPUs on this node
type GPUMetricsWithProm map[string]*scraper.GPUMetrics

func (g *GPUMetricsWithProm) Clone() framework.StateData {
	res := make(GPUMetricsWithProm)
	for k, v := range *g {
		res[k] = v.Clone()
	}
	return &res
}
