package schedule

import (
	"context"
	"genius/pkg/monitor"
	"genius/pkg/schedule/filter"
	"genius/pkg/schedule/score"
	"genius/pkg/schedule/sort"
	"genius/pkg/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sync"
)

const (
	SchedulerName = "genius"
)

const (
	metricsKey = "metrics"
)

var (
	_ framework.QueueSortPlugin = &Genius{}
	_ framework.PreFilterPlugin = &Genius{}
	_ framework.FilterPlugin    = &Genius{}
	_ framework.ScorePlugin     = &Genius{}
)

type Genius struct {
	handle  framework.Handle
	monitor *monitor.Monitor
	*sync.RWMutex
}

func New(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	m, err := monitor.NewMonitor(monitor.Scheme, monitor.PromHost, monitor.PromPort)
	if err != nil {
		klog.Exitf("creating gpu monitor error: %v", err)
	}

	return &Genius{
		handle:  handle,
		monitor: m,
	}, nil
}

func (g *Genius) Name() string {
	return SchedulerName
}

func (g *Genius) Less(podInfo1, podInfo2 *framework.QueuedPodInfo) bool {
	return sort.Less(podInfo1, podInfo2)
}

func (g *Genius) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	klog.V(3).Infof("prefilter pod %v, updating metrics for next scheduling phases", pod.Name)

	metrics, err := g.monitor.UpdateMetrics()
	if err != nil {
		klog.Errorf("updating metrics for scheduling error: %v", err)
		return framework.NewStatus(framework.Error)
	}

	state.Lock()
	defer state.Unlock()
	state.Write(metricsKey, metrics)
	return framework.NewStatus(framework.Success)
}

func (g *Genius) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (g *Genius) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.V(3).Infof("filter pod %v and node %v", pod.Name, nodeInfo.Node().Name)

	g.RLock()
	metrics, err := state.Read(metricsKey)
	g.RUnlock()

	if err != nil {
		klog.Errorf("retrieving cluster metrics from cyclestate in filter phase error: %v", err)
		return framework.NewStatus(framework.Error, "cannot retrieve cluster metrics")
	}

	m := metrics.(*types.GPUMetricsWithProm)
	if ok, requiredNumber := filter.PodFitsGPUNumber(pod, nodeInfo, m); ok {
		fitsMemoryEach := filter.PodFitsMemoryEach(requiredNumber, pod, nodeInfo, m)
		fitsMemoryTotal := filter.PodFitsMemoryTotal(pod, nodeInfo, m)
		fitsModel := filter.PodFitsModel(requiredNumber, pod, nodeInfo, m)
		if fitsMemoryEach && fitsMemoryTotal && fitsModel {
			return framework.NewStatus(framework.Success)
		}
	}

	return framework.NewStatus(framework.Unschedulable, "unschedulable node: "+nodeInfo.Node().Name)
}

func (g *Genius) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.V(3).Infof("scoring pod %v and node %v", pod.Name, nodeName)

	nodeInfo, err := g.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		klog.Errorf("getting node info error: %v", err)
		return 0, framework.NewStatus(framework.Error)
	}

	g.RLock()
	metrics, err := state.Read(metricsKey)
	g.RUnlock()
	if err != nil {
		klog.Errorf("retrieving cluster metrics from cyclestate in scoring phase error: %v", err)
		return 0, framework.NewStatus(framework.Error)
	}

	m := metrics.(*types.GPUMetricsWithProm)
	sc, err := score.ComputeScore(pod, nodeInfo, m)
	if err != nil {
		klog.Errorf("computing score of pod %v and node %v error: %v", pod.Name, nodeName, err)
		return 0, framework.NewStatus(framework.Error)
	}

	return int64(sc), nil
}

func (g *Genius) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {

	return framework.NewStatus(framework.Success)
}

func (g *Genius) ScoreExtensions() framework.ScoreExtensions {
	return nil
}
