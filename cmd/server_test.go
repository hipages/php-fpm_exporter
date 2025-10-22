package cmd

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestParseConstLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected prometheus.Labels
		err      string
	}{
		{
			name:     "empty input",
			input:    nil,
			expected: nil,
		},
		{
			name:  "single label",
			input: []string{"env=prod"},
			expected: prometheus.Labels{
				"env": "prod",
			},
		},
		{
			name:  "multiple labels with commas",
			input: []string{"team=core,service=api", "env=prod"},
			expected: prometheus.Labels{
				"team":    "core",
				"service": "api",
				"env":     "prod",
			},
		},
		{
			name:  "duplicate label key",
			input: []string{"env=prod", "env=staging"},
			err:   `duplicate label key "env"`,
		},
		{
			name:  "missing key",
			input: []string{"=prod"},
			err:   "label key cannot be empty",
		},
		{
			name:  "missing equals",
			input: []string{"env"},
			err:   "expected key=value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			labels, err := parseConstLabels(tc.input)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				require.Nil(t, labels)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, labels)
		})
	}
}

func TestGoMetricsIncludeConstLabels(t *testing.T) {
	constLabels := prometheus.Labels{"const": "value"}

	dummy := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "dummy_metric",
		Help:        "dummy metric for registry setup",
		ConstLabels: constLabels,
	})
	dummy.Set(1)

	registry, _ := newMetricsRegistry(constLabels, dummy)

	families, err := registry.Gather()
	require.NoError(t, err)

	var dummyHasConst, goHasConst bool
	for _, family := range families {
		switch {
		case family.GetName() == "dummy_metric":
			dummyHasConst = familyHasLabel(family, "const", "value")
		case strings.HasPrefix(family.GetName(), "go_"):
			goHasConst = goHasConst || familyHasLabel(family, "const", "value")
		}
	}

	require.True(t, dummyHasConst, "expected dummy_metric to include the constant label")
	require.True(t, goHasConst, "expected at least one go_* metric to include the constant label")
}

func metricHasLabel(metric *dto.Metric, name, value string) bool {
	for _, label := range metric.GetLabel() {
		if label.GetName() == name && label.GetValue() == value {
			return true
		}
	}

	return false
}

func familyHasLabel(family *dto.MetricFamily, name, value string) bool {
	for _, metric := range family.Metric {
		if metricHasLabel(metric, name, value) {
			return true
		}
	}

	return false
}
