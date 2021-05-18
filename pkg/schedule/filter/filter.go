package filter

import (
	"genius/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"regexp"
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
		if nInt <= gpuNumberOnThisNode {
			klog.Infof(`pod %v passed the gpu number filter successfully`, pod.Name)
			return true, nInt
		}

		klog.Infof(`pod %v does not passed the gpu number filter, since it requires %v gpu, but there are %v on this node`,
			pod.Name, number, gpuNumberOnThisNode)
		return false, nInt
	}
	klog.Infof(`pod %v does not specify the label "genius/gpu-number", skipping gpu number filter`, pod.Name)
	klog.Infof(`pod %v passed the gpu number filter successfully`, pod.Name)
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
		if fittedCards >= requiredNumber {
			klog.Infof(`pod %v passed the gpu memory-each filter successfully`, pod.Name)
			return true
		}

		klog.Infof(`pod %v does not pass the gpu memory-each filter, since it requires %v memory on each gpu, but only %v/%v gpu could satisfy`,
			pod.Name, memory, fittedCards, requiredNumber)
		return false
	}
	klog.Infof(`pod %v passed the gpu memory-each filter successfully`, pod.Name)
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
		if memoryInt <= totalMemory {
			klog.Infof("pod %v passed the gpu memory-total filter successfully", pod.Name)
			return true
		}

		klog.Infof(`pod %v does not pass the gpu memory-total filter, since it requires total %v memory, but the actual gpu memory in total is %v on this node`,
			pod.Name, memory, totalMemory)
		return false
	}
	klog.Infof("pod %v passed the gpu memory-total filter successfully", pod.Name)
	return true
}

// PodFitsModel judges whether there is enough number of cards in the model satisfies the number of cards
// required by the user in the specific same model.
// TODO: This filter-point can be more fine-grained. Maybe to specify the number of cards in the model makes
// more sense, but I'm not yet quite sure.
func PodFitsModel(requiredNumber int, pod *v1.Pod, nodeInfo *framework.NodeInfo, metrics *types.GPUMetricsWithProm) bool {
	gpus := (*metrics)[nodeInfo.Node().Name].GPUs
	fittedCards := 0
	if model, ok := pod.GetLabels()["genius/gpu-model"]; ok {
		for _, gpu := range gpus {
			if matchModel(model, gpu.StaticAttr.Model) {
				fittedCards++
			}
		}
		if fittedCards >= requiredNumber {
			klog.Infof(`pod %v passed the gpu model filter successfully`, pod.Name)
			return true
		}

		klog.Infof(`pod %v does not pass the gpu model filter, since it requires %v gpu of model "%v", but only %v/%v gpu could satisfy`,
			pod.Name, requiredNumber, model, fittedCards, requiredNumber)
		return false
	}
	klog.Infof(`pod %v passed the gpu model filter successfully`, pod.Name)
	return true
}

func matchModel(origin, request string) bool {
	r, _ := regexp.MatchString(strings.ToLower(origin), strings.ToLower(request))
	return r
}

func str2Int(s string) int {
	res, _ := strconv.Atoi(s)
	return res
}

func str2UInt64(s string) uint64 {
	val, _ := strconv.ParseUint(s, 10, 64)
	return val
}
