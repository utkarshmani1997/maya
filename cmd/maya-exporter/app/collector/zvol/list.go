package zvol

import (
	"sync"
	"time"

	"github.com/golang/glog"
	zvol "github.com/openebs/maya/pkg/zvol/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// volumeList implements prometheus.Collector interface
type volumeList struct {
	sync.Mutex
	listMetrics
	request bool
}

// NewVolumeList returns new instance of volumeList
func NewVolumeList() *volumeList {
	return &volumeList{
		listMetrics: newListMetrics(),
	}
}

func (v *volumeList) isRequestInProgress() bool {
	return v.request
}

// collectors returns the list of the collectors
func (v *volumeList) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		v.used,
		v.available,
		v.zfsParseErrorCounter,
		v.zfsListCommandErrorCounter,
		v.zfsListRequestRejectCounter,
	}
}

// gaugeVec returns list of zfs Gauge vectors (prometheus's type)
// in which values will be set.
// NOTE: Please donot edit the order, add new metrics at the end
// of the list
func (v *volumeList) gaugeVec() []*prometheus.GaugeVec {
	return []*prometheus.GaugeVec{
		v.used,
		v.available,
	}
}

// Describe is implementation of Describe method of prometheus.Collector
// interface.
func (v *volumeList) Describe(ch chan<- *prometheus.Desc) {
	for _, col := range v.collectors() {
		col.Describe(ch)
	}
}

func (v *volumeList) get() ([]fields, error) {
	v.Lock()
	defer v.Unlock()
	var timeout = 5 * time.Second

	glog.V(2).Info("Run zfs list command")
	stdout, err := zvol.Run(timeout, runner, "list", "-Hp")
	if err != nil {
		v.zfsListCommandErrorCounter.Inc()
		return nil, err
	}

	glog.V(2).Infof("Parse stdout of zfs list command, got stdout: \n%v", string(stdout))
	list, err := listParser(stdout, &v.listMetrics)

	return list, err
}

// Collect is implementation of prometheus's prometheus.Collector interface
func (v *volumeList) Collect(ch chan<- prometheus.Metric) {
	v.Lock()
	if v.isRequestInProgress() {
		v.zfsListRequestRejectCounter.Inc()
		v.Unlock()
		return

	}

	v.request = true
	v.Unlock()

	volumeLists, err := v.get()
	if err != nil {
		return
	}

	glog.V(2).Infof("Got zfs list: %#v", volumeLists)
	v.setListStats(volumeLists)
	for _, col := range v.collectors() {
		col.Collect(ch)
	}

	v.Lock()
	v.request = false
	v.Unlock()
}

func (v *volumeList) setListStats(volumeLists []fields) {
	for _, vol := range volumeLists {
		v.used.WithLabelValues(vol.name).Set(vol.used)
		v.available.WithLabelValues(vol.name).Set(vol.available)
	}
}