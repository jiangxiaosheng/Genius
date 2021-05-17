package score

import "github.com/observerward/pkg/scraper"

const (
	usedMemoryWeight         = 2
	powerWeight              = 1
	encoderUtilizationWeight = 1
	decoderUtilizationWeight = 1
)

func computeDynamicScore(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	return (scoreAgainstUsedMemory(gpuMetrics) + scoreAgainstPower(gpuMetrics, aggregatedMetrics) +
		scoreAgainstDecoderUtilization(gpuMetrics) + scoreAgainstEncoderUtilization(gpuMetrics)) * 100
}

func scoreAgainstUsedMemory(gpuMetrics *scraper.GPUMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += float32(gpu.StaticAttr.MemorySizeMB-gpu.UsedGlobalMemory) / float32(gpu.StaticAttr.MemorySizeMB)
	}
	return score * usedMemoryWeight
}

func scoreAgainstPower(gpuMetrics *scraper.GPUMetrics, aggregatedMetrics *clusterAggregatedMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += float32(gpu.Power) / float32(aggregatedMetrics.dynamic.power) * float32(aggregatedMetrics.cardsCount)
	}
	return score * powerWeight
}

func scoreAgainstEncoderUtilization(gpuMetrics *scraper.GPUMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += float32(100-gpu.EncoderUtilization) / 100
	}
	return score * encoderUtilizationWeight
}

func scoreAgainstDecoderUtilization(gpuMetrics *scraper.GPUMetrics) float32 {
	score := float32(0)
	for _, gpu := range gpuMetrics.GPUs {
		score += float32(100-gpu.DecoderUtilization) / 100
	}
	return score * decoderUtilizationWeight
}
