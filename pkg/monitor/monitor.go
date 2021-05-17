package monitor

import (
	"context"
	"errors"
	"fmt"
	"genius/pkg/types"
	"github.com/observerward/pkg/scraper"
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
	PromHost = "0.0.0.1"
	PromPort = 30090
)

const (
	baseQueryFilterInner = `__name__=~"observerward_.*"`
	baseQueryFilter      = `{__name__=~"observerward_.*"}`
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
		klog.Exitf("Creating Prometheus Client Error: %v", err)
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
	//cardsCount := len(values) / types.MetricsTypesCount
	//for _, value := range values {
	//	t := types.ExtractMetricTypeFromProm(value)
	//	nodename := types.ExtractNodeNameFromProm(value)
	//
	//}
	//for i := 0; i < cardsCount; i++ {
	//	firstRecord := values[i*types.MetricsTypesCount]
	//	nodeName := types.ExtractNodeNameFromProm(firstRecord)
	//	gpuID := types.ExtractIDFromProm(firstRecord)
	//	uuid := types.ExtractUUIDFromProm(firstRecord)
	//	if gpuMetricsWithProm[nodeName] == nil {
	//		gpuMetricsWithProm[nodeName] = &scraper.GPUMetrics{}
	//	}
	//	newCard := &scraper.MetricsSnapshotPerGPU{}
	//	newCard.StaticAttr.ID = gpuID
	//	newCard.StaticAttr.UUID = uuid
	//	for idx := 0; idx < types.MetricsTypesCount; idx++ {
	//		tp := types.ExtractMetricTypeFromProm(values[i*types.MetricsTypesCount+idx])
	//		klog.Info(values[i*types.MetricsTypesCount+idx])
	//		val := types.ExtractValueFromProm(values[i*types.MetricsTypesCount+idx])
	//		klog.Infof("%v %v", tp, val)
	//		switch tp {
	//		case types.GPUDecoderUtilization:
	//			newCard.DecoderUtilization = uint(val)
	//		case types.GPUEncoderUtilization:
	//			newCard.EncoderUtilization = uint(val)
	//		case types.GPUMemoryUtilization:
	//			newCard.MemoryUtilization = uint(val)
	//		case types.GPUPowerUsage:
	//			newCard.Power = uint(val)
	//		case types.GPUUsedGlobalMemory:
	//			newCard.UsedGlobalMemory = val
	//		case types.GPUMemorySize:
	//			newCard.StaticAttr.MemorySizeMB = val
	//		case types.GPUMultiprocessorCount:
	//			newCard.StaticAttr.MultiprocessorCount = uint32(val)
	//		case types.GPUSharedDecoderCount:
	//			newCard.StaticAttr.SharedDecoderCount = uint32(val)
	//		case types.GPUSharedEncoderCount:
	//			newCard.StaticAttr.SharedEncoderCount = uint32(val)
	//		}
	//	}
	//	gpuMetricsWithProm[nodeName].GPUs = append(gpuMetricsWithProm[nodeName].GPUs, newCard)
	//}
	return &metricsWithProm, nil
}

func (m *Monitor) queryLabelValues(labelname string, matchers []string) (model.LabelValues, error) {
	v1api := v1.NewAPI(m.client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	values, warning, err := v1api.LabelValues(ctx, labelname, matchers, time.Now().Add(-time.Hour), time.Now())

	for _, w := range warning {
		klog.Warningf("Warning on Querying Label Values: %v\n", w)
	}

	if err != nil {
		klog.Errorf("Error on Querying Label Value: %v\n", err)
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
		klog.Errorf("Querying Prometheus Error: %v", err)
		return "", err
	}
	if len(warnings) > 0 {
		klog.Warningf("Warnings: %v", warnings)
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
