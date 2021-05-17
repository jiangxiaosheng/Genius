package filter

import (
	"genius/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"strconv"
	"strings"
)

// PodFitsGPUNumber judges whether the number of gpus on this node satisfies
// the required number specified in the label.
// If there is not such an "genius/gpu-number" label while there are gpu/gpus
// on this node, this function returns true, otherwise false.
func PodFitsGPUNumber(pod *v1.Pod, nodeInfo *framework.NodeInfo, metrics *types.GPUMetricsWithProm) (bool, int) {
	gpus := (*metrics)[nodeInfo.Node().Name].GPUs
	gpuNumberOnThisNode := len(gpus)
	if number, ok := pod.GetLabels()["genius/gpu-number"]; ok {
		nInt := str2Int(number)
		return nInt <= gpuNumberOnThisNode, nInt
	}
	return gpuNumberOnThisNode > 0, 0
}

// PodFitsMemoryEach judges whether each GPU on this node satisfies the memory
// requirement of the pod. However, this is a coarse-grained implementation, which
// means that the "genius/gpu-memory-each" label specifies the memory requirement that each
// GPU must satisfy. If any of the GPU does not have so much memory, then this
// function returns false.
func PodFitsMemoryEach(requiredNumber int, pod *v1.Pod, nodeInfo *framework.NodeInfo, metrics *types.GPUMetricsWithProm) bool {
	gpus := (*metrics)[nodeInfo.Node().Name].GPUs
	fittedCards := 0
	if memory, ok := pod.GetLabels()["genius/gpu-memory-each"]; ok {
		memoryInt := str2UInt64(memory)
		for _, gpu := range gpus {
			if gpu.FreeGlobalMemory > memoryInt {
				fittedCards++
			}
		}
		return fittedCards >= requiredNumber
	}
	return true
}

// PodFitsMemoryTotal judges whether the total GPU memory on this node satisfies
// the one required by the pod, which is specified through the "genius/gpu-memory-total" label.
// It does the comparison by aggregating the free global memory of each GPU on this node.
func PodFitsMemoryTotal(pod *v1.Pod, nodeInfo *framework.NodeInfo, metrics *types.GPUMetricsWithProm) bool {
	gpus := (*metrics)[nodeInfo.Node().Name].GPUs
	totalMemory := uint64(0)
	if memory, ok := pod.GetLabels()["genius/gpu-memory-total"]; ok {
		memoryInt := str2UInt64(memory)
		for _, gpu := range gpus {
			totalMemory += gpu.FreeGlobalMemory
		}
		return memoryInt <= totalMemory
	}
	return true
}

// PodFitsModel judges whether there is enough number of cards in the model satisfies the number of cards
// required by the user in the specific same model.
// TODO: This filter-point can be more fine-grained. Maybe to specify the number of cards in the model makes
// more sense, but I'm not yet quite sure.
func PodFitsModel(requiredNumber int, pod *v1.Pod, nodeInfo *framework.NodeInfo, metrics *types.GPUMetricsWithProm) bool {
	gpus := (*metrics)[nodeInfo.Node().Name].GPUs
	fittedCards := 0
	if model, ok := pod.GetLabels()["genius/model"]; ok {
		for _, gpu := range gpus {
			if strings.ToLower(model) == strings.ToLower(gpu.StaticAttr.Model) {
				fittedCards++
			}
		}
		return fittedCards >= requiredNumber
	}
	return true
}

func str2Int(s string) int {
	res, _ := strconv.Atoi(s)
	return res
}

func str2UInt64(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}
