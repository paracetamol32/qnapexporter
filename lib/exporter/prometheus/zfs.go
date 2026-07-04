package prometheus

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/pedropombeiro/qnapexporter/lib/utils"
)

type zfsArcMetricDef struct {
	name       string
	keys       []string
	help       string
	metricType string
}

var zfsArcMetricDefs = []zfsArcMetricDef{
	{
		name:       "qnap_zfs_arc_size_bytes",
		keys:       []string{"size"},
		help:       "Current ARC size in bytes",
		metricType: "gauge",
	},
	{
		name:       "qnap_zfs_arc_target_size_bytes",
		keys:       []string{"c"},
		help:       "Target ARC size in bytes",
		metricType: "gauge",
	},
	{
		name:       "qnap_zfs_arc_min_size_bytes",
		keys:       []string{"c_min"},
		help:       "Minimum ARC size in bytes",
		metricType: "gauge",
	},
	{
		name:       "qnap_zfs_arc_max_size_bytes",
		keys:       []string{"c_max"},
		help:       "Maximum ARC size in bytes",
		metricType: "gauge",
	},
	{
		name:       "qnap_zfs_arc_hits_total",
		keys:       []string{"hits"},
		help:       "Total number of ARC hits",
		metricType: "counter",
	},
	{
		name:       "qnap_zfs_arc_misses_total",
		keys:       []string{"misses"},
		help:       "Total number of ARC misses",
		metricType: "counter",
	},
	{
		name:       "qnap_zfs_l2arc_size_bytes",
		keys:       []string{"l2_size"},
		help:       "Current L2ARC size in bytes",
		metricType: "gauge",
	},
	{
		name:       "qnap_zfs_l2arc_allocated_bytes",
		keys:       []string{"l2_asize", "l2_allocated"},
		help:       "Allocated L2ARC size in bytes",
		metricType: "gauge",
	},
	{
		name:       "qnap_zfs_l2arc_hits_total",
		keys:       []string{"l2_hits"},
		help:       "Total number of L2ARC hits",
		metricType: "counter",
	},
	{
		name:       "qnap_zfs_l2arc_misses_total",
		keys:       []string{"l2_misses"},
		help:       "Total number of L2ARC misses",
		metricType: "counter",
	},
	{
		name:       "qnap_zfs_l2arc_read_bytes_total",
		keys:       []string{"l2_read_bytes"},
		help:       "Total number of bytes read from L2ARC",
		metricType: "counter",
	},
	{
		name:       "qnap_zfs_l2arc_write_bytes_total",
		keys:       []string{"l2_write_bytes"},
		help:       "Total number of bytes written to L2ARC",
		metricType: "counter",
	},
	{
		name:       "qnap_zfs_l2arc_feeds_total",
		keys:       []string{"l2_feeds"},
		help:       "Total number of L2ARC feed attempts",
		metricType: "counter",
	},
}

func (e *promExporter) getZFSArcStatsMetrics() ([]metric, error) {
	fi, err := os.Stat(e.zfsArcstats)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	values := map[string]string{}

	if fi.IsDir() {
		for _, def := range zfsArcMetricDefs {
			for _, key := range def.keys {
				if _, ok := values[key]; ok {
					continue
				}

				b, err := os.ReadFile(filepath.Join(e.zfsArcstats, key))
				if err != nil {
					continue
				}

				values[key] = strings.TrimSpace(string(b))
			}
		}
	} else {
		lines, err := utils.ReadFileLines(e.zfsArcstats)
		if err != nil {
			return nil, err
		}

		values = parseZFSArcStats(lines)
	}
	metrics := make([]metric, 0, len(zfsArcMetricDefs))
	for _, def := range zfsArcMetricDefs {
		valueStr, ok := firstZFSArcStatValue(values, def.keys)
		if !ok {
			continue
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parse ZFS arcstats %s: %w", def.keys[0], err)
		}

		metrics = append(metrics, metric{
			name:       def.name,
			value:      value,
			help:       def.help,
			metricType: def.metricType,
		})
	}

	return metrics, nil
}

func parseZFSArcStats(lines []string) map[string]string {
	values := make(map[string]string, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		values[fields[0]] = fields[2]
	}

	return values
}

func firstZFSArcStatValue(values map[string]string, keys []string) (string, bool) {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value, true
		}
	}

	return "", false
}
