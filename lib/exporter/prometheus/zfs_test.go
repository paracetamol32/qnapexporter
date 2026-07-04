package prometheus

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetZFSArcStatsMetrics(t *testing.T) {
	arcstats := `13 1 0x01 13 624 123456789 987654321
name                            type data
hits                            4    1001
misses                          4    202
size                            4    123456
c                               4    234567
c_min                           4    345678
c_max                           4    456789
l2_size                         4    567890
l2_asize                        4    678901
l2_hits                         4    303
l2_misses                       4    404
l2_read_bytes                   4    5050
l2_write_bytes                  4    6060
l2_feeds                        4    707`

	path := filepath.Join(t.TempDir(), "arcstats")
	require.NoError(t, os.WriteFile(path, []byte(arcstats), 0644))

	e := &promExporter{
		ExporterConfig: ExporterConfig{Logger: log.New(os.Stderr, "", 0)},
		zfsArcstats:    path,
	}

	metrics, err := e.getZFSArcStatsMetrics()
	require.NoError(t, err)
	require.Len(t, metrics, len(zfsArcMetricDefs))

	assertMetricValue(t, metrics, "qnap_zfs_arc_size_bytes", 123456)
	assertMetricValue(t, metrics, "qnap_zfs_arc_target_size_bytes", 234567)
	assertMetricValue(t, metrics, "qnap_zfs_arc_min_size_bytes", 345678)
	assertMetricValue(t, metrics, "qnap_zfs_arc_max_size_bytes", 456789)
	assertMetricValue(t, metrics, "qnap_zfs_arc_hits_total", 1001)
	assertMetricValue(t, metrics, "qnap_zfs_arc_misses_total", 202)
	assertMetricValue(t, metrics, "qnap_zfs_l2arc_size_bytes", 567890)
	assertMetricValue(t, metrics, "qnap_zfs_l2arc_allocated_bytes", 678901)
	assertMetricValue(t, metrics, "qnap_zfs_l2arc_hits_total", 303)
	assertMetricValue(t, metrics, "qnap_zfs_l2arc_misses_total", 404)
	assertMetricValue(t, metrics, "qnap_zfs_l2arc_read_bytes_total", 5050)
	assertMetricValue(t, metrics, "qnap_zfs_l2arc_write_bytes_total", 6060)
	assertMetricValue(t, metrics, "qnap_zfs_l2arc_feeds_total", 707)

	for _, m := range metrics {
		assert.NotEmpty(t, m.help)
		assert.Contains(t, []string{"counter", "gauge"}, m.metricType)
	}
}

func TestGetZFSArcStatsMetricsMissingFile(t *testing.T) {
	e := &promExporter{zfsArcstats: filepath.Join(t.TempDir(), "missing")}

	metrics, err := e.getZFSArcStatsMetrics()
	require.NoError(t, err)
	assert.Empty(t, metrics)
}

func TestGetZFSArcStatsMetricsUsesL2AllocatedFallback(t *testing.T) {
	arcstats := `name type data
l2_allocated 4 1234`

	path := filepath.Join(t.TempDir(), "arcstats")
	require.NoError(t, os.WriteFile(path, []byte(arcstats), 0644))

	e := &promExporter{zfsArcstats: path}

	metrics, err := e.getZFSArcStatsMetrics()
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	assert.Equal(t, "qnap_zfs_l2arc_allocated_bytes", metrics[0].name)
	assert.Equal(t, float64(1234), metrics[0].value)
}

func TestGetZFSArcStatsMetricsInvalidValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arcstats")
	require.NoError(t, os.WriteFile(path, []byte("hits 4 nope"), 0644))

	e := &promExporter{zfsArcstats: path}

	_, err := e.getZFSArcStatsMetrics()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse ZFS arcstats hits")
}

func assertMetricValue(t *testing.T, metrics []metric, name string, expected float64) {
	t.Helper()

	for _, m := range metrics {
		if m.name == name {
			assert.Equal(t, expected, m.value)
			return
		}
	}

	t.Fatalf("metric %q not found", name)
}
