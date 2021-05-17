package monitor

import (
	"github.com/prometheus/common/log"
	"testing"
)

var m *Monitor

func init() {
	var err error
	m, err = NewMonitor("http", "localhost", 30090)
	if err != nil {
		log.Error(err)
		return
	}
}

func TestValue2String(t *testing.T) {
	val, err := m.query(`{__name__=~"observerward_.*"}`)
	if err != nil {
		log.Error(err)
		return
	}
	println(val)
}

func TestQueryLabels(t *testing.T) {
	vals, err := m.queryLabelValues(`kubernetes_node`, nil)
	if err != nil {
		log.Error(err)
		return
	}
	for _, v := range vals {
		log.Infoln(v)
	}
}

func TestQueryByLabel(t *testing.T) {
	m.queryByLabel([]string{"id"}, []string{"0"})
}

func TestGenerateFilters(t *testing.T) {
	labelnames := []string{"id", "uuid"}
	labelvalues := []string{"0", "0000"}
	log.Infoln(generateFilters(labelnames, labelvalues))
}

func TestUpdateMetrics(t *testing.T) {
	metrics, err := m.UpdateMetrics()
	if err != nil {
		log.Error(err)
		return
	}
	for k, v := range *metrics {
		log.Infof("NodeName: %v", k)
		for _, g := range v.GPUs {
			log.Infof(" ID: %v", g.StaticAttr.ID)
			log.Infof(" UUID: %v", g.StaticAttr.UUID)
			log.Infof(" model: %v", g.StaticAttr.Model)
			log.Infof(" decoder utilization: %v", g.DecoderUtilization)
			log.Infof(" encoder utilization: %v", g.EncoderUtilization)
			log.Infof(" memory utilization: %v", g.MemoryUtilization)
			log.Infof(" power usage: %v", g.Power)
			log.Infof(" used global memory: %v", g.UsedGlobalMemory)
			log.Infof(" free global memory: %v", g.FreeGlobalMemory)
			log.Infof(" memory size in MB: %v", g.StaticAttr.MemorySizeMB)
			log.Infof(" multiprocessor count: %v", g.StaticAttr.MultiprocessorCount)
			log.Infof(" shared decoder count: %v", g.StaticAttr.SharedDecoderCount)
			log.Infof(" shared encoder count: %v", g.StaticAttr.SharedEncoderCount)
		}
	}
}
