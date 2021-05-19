package score

import (
	"github.com/observerward/pkg/scraper"
)

const (
	memoryWeight             = 2
	multiprocessorWeight     = 2
	sharedDecoderCountWeight = 1
	sharedEncoderCountWeight = 1
	bandwidthWeight          = 2
)

type staticMetricsOnNode staticMetrics

func computeStaticScore(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	smn := &staticMetricsOnNode{}
	for _, gpu := range gpuMetrics.GPUs {
		smn.memorySize += gpu.StaticAttr.MemorySizeMB
		smn.bandwidth += uint64(gpu.StaticAttr.Bandwidth)
		smn.multiprocessorCount += uint64(gpu.StaticAttr.MultiprocessorCount)
		smn.sharedDecoderCount += uint64(gpu.StaticAttr.SharedDecoderCount)
		smn.sharedEncoderCount += uint64(gpu.StaticAttr.SharedEncoderCount)
	}
	return scoreAgainstMemory(smn, aggregatedMetrics) + scoreAgainstMultiprocessor(smn, aggregatedMetrics) +
		scoreAgainstSharedDecoder(smn, aggregatedMetrics) + scoreAgainstSharedEncoder(smn, aggregatedMetrics) +
		scoreAgainstBandwidth(smn, aggregatedMetrics)
}

func scoreAgainstMemory(metricsOnNode *staticMetricsOnNode, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	return float32(metricsOnNode.memorySize) / float32(aggregatedMetrics.static.memorySize) * float32(aggregatedMetrics.cardsCount) * memoryWeight
}

func scoreAgainstMultiprocessor(metricsOnNode *staticMetricsOnNode, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	return float32(metricsOnNode.multiprocessorCount) / float32(aggregatedMetrics.static.multiprocessorCount) * float32(aggregatedMetrics.cardsCount) * multiprocessorWeight
}

func scoreAgainstSharedDecoder(metricsOnNode *staticMetricsOnNode, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	return float32(metricsOnNode.sharedDecoderCount) / float32(aggregatedMetrics.static.sharedDecoderCount) * float32(aggregatedMetrics.cardsCount) * sharedDecoderCountWeight
}

func scoreAgainstSharedEncoder(metricsOnNode *staticMetricsOnNode, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	return float32(metricsOnNode.sharedEncoderCount) / float32(aggregatedMetrics.static.sharedEncoderCount) * float32(aggregatedMetrics.cardsCount) * sharedEncoderCountWeight
}

func scoreAgainstBandwidth(metricsOnNode *staticMetricsOnNode, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	return float32(metricsOnNode.bandwidth) / float32(aggregatedMetrics.static.bandwidth) * float32(aggregatedMetrics.cardsCount) * bandwidthWeight
}
