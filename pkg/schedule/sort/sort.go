package sort

import (
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"strconv"
)

func Less(podInfo1, podInfo2 *framework.QueuedPodInfo) bool {
	return priority(podInfo1) > priority(podInfo2)
}

func priority(podInfo *framework.QueuedPodInfo) int {
	if p, ok := podInfo.Pod.Labels["genius/priority"]; ok {
		pInt, _ := strconv.Atoi(p)
		return pInt
	}
	return 0
}
