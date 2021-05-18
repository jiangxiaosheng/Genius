package monitor

import (
	"context"
	"errors"
	"fmt"
	"github.com/genius/pkg/types"
	"github.com/jiangxiaosheng/ObserverWard/"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog/v2"
)

var (
	Scheme   = "http"
	PromHost = "222.201.144.187" // fixme: it should be discovered in runtime by k8s client sdk
	PromPort = 30090
)

const (
	baseQueryFilterInner = `__name__=~"observerward_.*"`
	k8sNodeNameLabel     = "kubernetes_node"
	idLabel              = "id"
)

type Monitor struct {
	PromAddress string
	client      api.Client
}

// NewMonitor returns a new monitor instance.
// The scheme parameter can be http or https.
// The host and port parameter specify the remote host and host
// which the prometheus http service is listen on.
func NewMonitor(scheme, host string, port int) (*Monitor, error) {
	address := strings.ToLower(scheme) + "://" + host + ":" + strconv.Itoa(port)
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		klog.Exitf("creating prometheus client error: %v", err)
	}

	return &Monitor{
		PromAddress: address,
		client:      client,
	}, nil
}

// UpdateMetrics returns the updated GPU metrics mapped by their nodename.
//
func (m *Monitor) UpdateMetrics() (*types.GPUMetricsWithProm, error) {
	nodenames, err := m.queryLabelValues(k8sNodeNameLabel, nil)
	if err != nil {
		klog.Errorf("querying kubernetes node names error: %v", err)
		return nil, err
	}
	metricsWithProm := make(types.GPUMetricsWithProm)
	for _, nodenameValue := range nodenames {
		nodename := string(nodenameValue)
		metricsWithProm[nodename] = &scraper.GPUMetrics{}

		idFilter, err := generateFilters([]string{k8sNodeNameLabel}, []string{nodename})
		numGPUValue, err := m.queryLabelValues(idLabel, []string{idFilter})
		if err != nil {
			klog.Errorf(`querying "id" label values on kubernetes node %v error: %v`, nodename, err)
			return nil, err
		}

		for _, i := range numGPUValue {
			id, err := strconv.Atoi(string(i))
			allFilter, err := generateFilters([]string{k8sNodeNameLabel, idLabel}, []string{nodename, string(i)})
			recordsStr, err := m.query(allFilter)
			if err != nil {
				klog.Errorf(`querying prometheus records of kubernetes node %v and GPU with id "%v" error: `,
					nodename, id, err)
				return nil, err
			}

			records := strings.Split(recordsStr, "\n")
			if len(records) != types.MetricsTypesCount {
				klog.Warningf(`the number of metric types from prometheus query results is invalid, current
				number is %v, but it should be %v`, len(records), types.MetricsTypesCount)
			}

			gpuSnapshot := &scraper.MetricsSnapshotPerGPU{}
			uuid := types.ExtractUUIDFromProm(records[0])
			gpuSnapshot.StaticAttr.UUID = uuid
			gpuSnapshot.StaticAttr.ID = uint(id)
			md := types.ExtractModelFromProm(records[0])
			gpuSnapshot.StaticAttr.Model = md
			for _, record := range records {
				t := types.ExtractMetricTypeFromProm(record)
				val := types.ExtractValueFromProm(record)
				switch t {
				case types.GPUDecoderUtilization:
					gpuSnapshot.DecoderUtilization = uint(val)
				case types.GPUEncoderUtilization:
					gpuSnapshot.EncoderUtilization = uint(val)
				case types.GPUFreeGlobalMemory:
					gpuSnapshot.FreeGlobalMemory = val
				case types.GPUMemoryUtilization:
					gpuSnapshot.MemoryUtilization = uint(val)
				case types.GPUPowerUsage:
					gpuSnapshot.Power = uint(val)
				case types.GPUUsedGlobalMemory:
					gpuSnapshot.UsedGlobalMemory = val
				case types.GPUMemorySize:
					gpuSnapshot.StaticAttr.MemorySizeMB = val
				case types.GPUMultiprocessorCount:
					gpuSnapshot.StaticAttr.MultiprocessorCount = uint32(val)
				case types.GPUSharedDecoderCount:
					gpuSnapshot.StaticAttr.SharedDecoderCount = uint32(val)
				case types.GPUSharedEncoderCount:
					gpuSnapshot.EncoderUtilization = uint(val)
				}
			}

			metricsWithProm[nodename].GPUs = append(metricsWithProm[nodename].GPUs, gpuSnapshot)
		}
	}
	return &metricsWithProm, nil
}

func (m *Monitor) queryLabelValues(labelname string, matchers []string) (model.LabelValues, error) {
	v1api := v1.NewAPI(m.client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	values, warning, err := v1api.LabelValues(ctx, labelname, matchers, time.Now().Add(-time.Hour), time.Now())

	for _, w := range warning {
		klog.Warningf("warning on querying label values: %v\n", w)
	}

	if err != nil {
		klog.Errorf("error on querying label value: %v\n", err)
		return nil, err
	}
	return values, nil
}

func (m *Monitor) queryByLabel(labelnames, labelvalues []string) (string, error) {
	filters, err := generateFilters(labelnames, labelvalues)
	if err != nil {
		klog.Error(err)
		return "", err
	}

	values, err := m.query(filters)
	return values, nil
}

// query calls prometheus HTTP api to retrieve metrics.
// This function refers to the Instant queries on page https://prometheus.io/docs/prometheus/latest/querying/api/.
// The return value should be further processed against concrete business logic.
func (m *Monitor) query(qString string) (string, error) {
	v1api := v1.NewAPI(m.client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.Query(ctx, qString, time.Now())
	if err != nil {
		klog.Errorf("querying prometheus error: %v", err)
		return "", err
	}
	if len(warnings) > 0 {
		klog.Warningf("warnings: %v", warnings)
	}
	return value2String(&result), nil
}

func generateFilters(labelnames, labelvalues []string) (string, error) {
	if len(labelnames) != len(labelvalues) {
		return "", errors.New("length of labelnames don't match length of labelvalues")
	}

	filters := baseQueryFilterInner
	for i := 0; i < len(labelvalues); i++ {
		filters = fmt.Sprintf(`%v, %v="%v"`, filters, labelnames[i], labelvalues[i])
	}
	filters = fmt.Sprintf(`{%v}`, filters)
	return filters, nil
}

func value2String(value *model.Value) string {
	return fmt.Sprintf("%v", *value)
}
