package score

import (
	"github.com/genius/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type clusterAggregatedMetrics struct {
	cardsCount uint
	static     staticMetrics
	dynamic    dynamicMetrics
}

type staticMetrics struct {
	memorySize          uint64 // in MB
	multiprocessorCount uint64
	sharedDecoderCount  uint64
	sharedEncoderCount  uint64
	bandwidth           uint64
}

type dynamicMetrics struct {
	usedGlobalMemory   uint64
	freeGlobalMemory   uint64
	power              uint
	encoderUtilization uint
	decoderUtilization uint
	memoryUtilization  uint
}

const (
	dynamicWeight = 2
	staticWeight  = 1
)

func ComputeScore(pod *v1.Pod, nodeInfo *framework.NodeInfo, metrics *types.GPUMetricsWithProm) (uint64, error) {
	aggregatedMetrics := aggregateMetrics(metrics)
	staticScore := computeStaticScore((*metrics)[nodeInfo.Node().Name], aggregatedMetrics)
	dynamicScore := computeDynamicScore((*metrics)[nodeInfo.Node().Name], aggregatedMetrics)
	return uint64(staticScore*staticWeight + dynamicScore*dynamicWeight), nil
}

func aggregateMetrics(metrics *types.GPUMetricsWithProm) *clusterAggregatedMetrics {
	res := &clusterAggregatedMetrics{}
	for _, v := range *metrics {
		for _, gpu := range v.GPUs {
			res.cardsCount++
			res.static.memorySize += gpu.StaticAttr.MemorySizeMB
			res.static.bandwidth += uint64(gpu.StaticAttr.Bandwidth)
			res.static.multiprocessorCount += uint64(gpu.StaticAttr.MultiprocessorCount)
			res.static.sharedEncoderCount += uint64(gpu.StaticAttr.SharedEncoderCount)
			res.static.sharedDecoderCount += uint64(gpu.StaticAttr.SharedDecoderCount)

			res.dynamic.power += gpu.Power
			res.dynamic.encoderUtilization += gpu.EncoderUtilization
			res.dynamic.decoderUtilization += gpu.DecoderUtilization
			res.dynamic.memoryUtilization += gpu.MemoryUtilization
			res.dynamic.freeGlobalMemory += gpu.FreeGlobalMemory
			res.dynamic.usedGlobalMemory += gpu.UsedGlobalMemory
		}
	}
	return res
}
