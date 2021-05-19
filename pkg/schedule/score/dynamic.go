package score

import "github.com/observerward/pkg/scraper"

const (
	usedMemoryWeight         = 2
	powerWeight              = 1
	encoderUtilizationWeight = 1
	decoderUtilizationWeight = 1
)

func computeDynamicScore(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	return scoreAgainstFreeMemory(gpuMetrics, aggregatedMetrics) + scoreAgainstPower(gpuMetrics, aggregatedMetrics) +
		scoreAgainstDecoderUtilization(gpuMetrics, aggregatedMetrics) + scoreAgainstEncoderUtilization(gpuMetrics, aggregatedMetrics)
}

func scoreAgainstFreeMemory(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += float32(gpu.FreeGlobalMemory) / float32(aggregatedMetrics.dynamic.freeGlobalMemory) * float32(aggregatedMetrics.cardsCount)
	}
	return score * usedMemoryWeight / float32(len(gpuMetrics.GPUs))
}

func scoreAgainstPower(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += float32(gpu.Power) / float32(aggregatedMetrics.dynamic.power) * float32(aggregatedMetrics.cardsCount)
	}
	return score * powerWeight / float32(len(gpuMetrics.GPUs))
}

func scoreAgainstEncoderUtilization(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += (1 - float32(gpu.EncoderUtilization)) / float32(aggregatedMetrics.cardsCount-aggregatedMetrics.dynamic.encoderUtilization) *
			float32(aggregatedMetrics.cardsCount)
	}
	return score * encoderUtilizationWeight / float32(len(gpuMetrics.GPUs))
}

func scoreAgainstDecoderUtilization(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += (1 - float32(gpu.DecoderUtilization)) / float32(aggregatedMetrics.cardsCount-aggregatedMetrics.dynamic.decoderUtilization) *
			float32(aggregatedMetrics.cardsCount)
	}
	return score * decoderUtilizationWeight / float32(len(gpuMetrics.GPUs))
}
