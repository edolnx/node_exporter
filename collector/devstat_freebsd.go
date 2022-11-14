// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !nodevstat
// +build !nodevstat

package collector

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

// #cgo LDFLAGS: -ldevstat -lkvm -lelf
// #include "devstat_freebsd.h"
import "C"

const (
	devstatSubsystem = "devstat"
)

type devstatCollector struct {
	mu      sync.Mutex
	devinfo *C.struct_devinfo

	bytes        typedDesc
	transfers    typedDesc
	duration     typedDesc
	busyTime     typedDesc
	busy_percent typedDesc
	blocks       typedDesc
	tps          typedDesc
	mbps         typedDesc
	kbpt         typedDesc
	mspertxn     typedDesc
	queue_length typedDesc
	logger       log.Logger
}

func init() {
	registerCollector("devstat", defaultDisabled, NewDevstatCollector)
}

// NewDevstatCollector returns a new Collector exposing Device stats.
func NewDevstatCollector(logger log.Logger) (Collector, error) {
	return &devstatCollector{
		devinfo: &C.struct_devinfo{},
		bytes: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "bytes_total"),
			"The total number of bytes in transactions.",
			[]string{"device", "type"}, nil,
		), prometheus.CounterValue},
		transfers: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "transfers_total"),
			"The total number of transactions.",
			[]string{"device", "type"}, nil,
		), prometheus.CounterValue},
		duration: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "duration_seconds_total"),
			"The total duration of transactions in seconds.",
			[]string{"device", "type"}, nil,
		), prometheus.CounterValue},
		busyTime: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "busy_time_seconds_total"),
			"Total time the device had one or more transactions outstanding in seconds.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		busy_percent: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "busy_time_percentage_total"),
			"Total percentage of the block device time spent in busy.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		blocks: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "blocks_transferred_total"),
			"The total number of blocks transferred.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		queue_length: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "queue_length"),
			"The length of the command queue by device for pending operations.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		tps: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "transactions_per_second"),
			"The number of IO transactions per second for each device.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		mbps: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "megabytes_per_second"),
			"The throughput by operation type in megabytes per second for each device.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		kbpt: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "kilobytes_per_transfer"),
			"The average size of the transaction by operation type in kilobytes for each device.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		mspertxn: typedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, devstatSubsystem, "miliseconds_per_transaction"),
			"The average number of milliseconds per transaction per type for each device.",
			[]string{"device"}, nil,
		), prometheus.CounterValue},
		logger: logger,
	}, nil
}

func (c *devstatCollector) Update(ch chan<- prometheus.Metric) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var stats *C.Stats
	n := C._get_stats(c.devinfo, &stats)
	if n == -1 {
		return errors.New("devstat_getdevs failed")
	}

	base := unsafe.Pointer(stats)
	for i := C.int(0); i < n; i++ {
		offset := i * C.int(C.sizeof_Stats)
		stat := (*C.Stats)(unsafe.Pointer(uintptr(base) + uintptr(offset)))

		device := fmt.Sprintf("%s%d", C.GoString(&stat.device[0]), stat.unit)
		ch <- c.bytes.mustNewConstMetric(float64(stat.bytes.read), device, "read")
		ch <- c.bytes.mustNewConstMetric(float64(stat.bytes.write), device, "write")
		ch <- c.transfers.mustNewConstMetric(float64(stat.transfers.other), device, "other")
		ch <- c.transfers.mustNewConstMetric(float64(stat.transfers.read), device, "read")
		ch <- c.transfers.mustNewConstMetric(float64(stat.transfers.write), device, "write")
		ch <- c.duration.mustNewConstMetric(float64(stat.duration.other), device, "other")
		ch <- c.duration.mustNewConstMetric(float64(stat.duration.read), device, "read")
		ch <- c.duration.mustNewConstMetric(float64(stat.duration.write), device, "write")
		ch <- c.busyTime.mustNewConstMetric(float64(stat.busy_time), device)
		ch <- c.blocks.mustNewConstMetric(float64(stat.blocks), device)
		ch <- c.busy_percent.mustNewConstMetric(float64(stat.busy_percent), device)
		ch <- c.queue_length.mustNewConstMetric(float64(stat.queue_length), device)
		ch <- c.tps.mustNewConstMetric(float64(stat.tps.read), device, "read")
		ch <- c.tps.mustNewConstMetric(float64(stat.tps.write), device, "write")
		ch <- c.tps.mustNewConstMetric(float64(stat.tps.free), device, "free")
		ch <- c.tps.mustNewConstMetric(float64(stat.tps.other), device, "other")
		ch <- c.tps.mustNewConstMetric(float64(stat.tps.total), device, "total")
		ch <- c.mbps.mustNewConstMetric(float64(stat.mbps.read), device, "read")
		ch <- c.mbps.mustNewConstMetric(float64(stat.mbps.write), device, "write")
		ch <- c.kbpt.mustNewConstMetric(float64(stat.kbpt.read), device, "read")
		ch <- c.kbpt.mustNewConstMetric(float64(stat.kbpt.write), device, "write")
		ch <- c.kbpt.mustNewConstMetric(float64(stat.kbpt.free), device, "free")
		ch <- c.mspertxn.mustNewConstMetric(float64(stat.mspertxn.read), device, "read")
		ch <- c.mspertxn.mustNewConstMetric(float64(stat.mspertxn.write), device, "write")
		ch <- c.mspertxn.mustNewConstMetric(float64(stat.mspertxn.other), device, "other")
	}
	C.free(unsafe.Pointer(stats))
	return nil
}
