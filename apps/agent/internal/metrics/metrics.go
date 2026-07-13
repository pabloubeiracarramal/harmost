package metrics

import (
	"context"

	harmostv1 "github.com/harmost/proto/gen/harmost/v1"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func Collect(ctx context.Context) *harmostv1.SystemMetrics {
	m := &harmostv1.SystemMetrics{}

	if percents, err := cpu.PercentWithContext(ctx, 0, false); err == nil && len(percents) > 0 {
		m.CpuUsagePercent = float32(percents[0])
	}
	if v, err := mem.VirtualMemoryWithContext(ctx); err == nil {
		m.MemoryUsedBytes = int64(v.Used)
		m.MemoryTotalBytes = int64(v.Total)
	}
	if d, err := disk.UsageWithContext(ctx, "/"); err == nil {
		m.DiskUsedBytes = int64(d.Used)
		m.DiskTotalBytes = int64(d.Total)
	}
	return m
}
