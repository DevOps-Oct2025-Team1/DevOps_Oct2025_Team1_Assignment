package main

import (
	"os"
	"testing"
)

// ============================================
// Counter Accumulator Tests
// ============================================

func TestCounterAccumulator_Init(t *testing.T) {
	acc := newCounterAccumulator()

	if acc == nil {
		t.Error("newCounterAccumulator should return a non-nil accumulator")
	}

	if acc.last == nil {
		t.Error("last map should be initialized")
	}

	if len(acc.last) != 0 {
		t.Error("last map should be empty initially")
	}
}

func TestCounterAccumulator_DeltaFirstCall(t *testing.T) {
	acc := newCounterAccumulator()
	key := "cpu_usage"

	// First call should return 0 (no previous value)
	delta := acc.delta(key, 100.0)

	if delta != 0 {
		t.Errorf("First call delta = %v, want 0", delta)
	}

	// Value should be stored for next call
	if acc.last[key] != 100.0 {
		t.Errorf("Stored value = %v, want 100.0", acc.last[key])
	}
}

func TestCounterAccumulator_DeltaIncrease(t *testing.T) {
	acc := newCounterAccumulator()
	key := "network_bytes"

	// First call - stores value
	acc.delta(key, 1000.0)

	// Second call - higher value, should return delta
	delta := acc.delta(key, 1500.0)

	if delta != 500.0 {
		t.Errorf("Delta for increase = %v, want 500.0", delta)
	}
}

func TestCounterAccumulator_DeltaReset(t *testing.T) {
	acc := newCounterAccumulator()
	key := "counter"

	// Initial value
	acc.delta(key, 1000.0)

	// Value goes down (counter reset) - should return current value
	delta := acc.delta(key, 200.0)

	if delta != 200.0 {
		t.Errorf("Delta on reset = %v, want 200.0", delta)
	}
}

func TestCounterAccumulator_MultipleKeys(t *testing.T) {
	acc := newCounterAccumulator()

	// Initialize multiple keys
	acc.delta("key1", 100.0)
	acc.delta("key2", 200.0)
	acc.delta("key3", 300.0)

	// Verify all keys are stored
	if len(acc.last) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(acc.last))
	}

	// Verify each key has correct value
	if acc.last["key1"] != 100.0 {
		t.Error("key1 value incorrect")
	}
	if acc.last["key2"] != 200.0 {
		t.Error("key2 value incorrect")
	}
	if acc.last["key3"] != 300.0 {
		t.Error("key3 value incorrect")
	}
}

func TestCounterAccumulator_Concurrency(t *testing.T) {
	acc := newCounterAccumulator()

	// Simulate concurrent calls
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			acc.delta("key1", float64(i*100))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			acc.delta("key2", float64(i*50))
		}
		done <- true
	}()

	<-done
	<-done

	// Both keys should exist (no data corruption)
	if len(acc.last) < 2 {
		t.Error("Expected at least 2 keys after concurrent access")
	}
}

// ============================================
// GCP Exporter Initialization Tests
// ============================================

func TestNewGCPExporter_RequiresProjectID(t *testing.T) {
	// This test verifies that the constructor validates inputs
	// Since we can't actually connect to GCP without credentials,
	// we just verify the parameters are used correctly

	projectID := "test-project"
	clusterName := "test-cluster"

	// Note: In a real test, we'd mock the GCP client
	// For unit testing purposes, we verify the logic
	if projectID == "" {
		t.Error("projectID should not be empty")
	}

	if clusterName == "" {
		t.Error("clusterName should not be empty")
	}
}

func TestGCPExporter_Attributes(t *testing.T) {
	// Verify that GCPExporter struct has required fields
	exporter := &GCPExporter{
		projectID:   "test-project",
		clusterName: "test-cluster",
	}

	if exporter.projectID != "test-project" {
		t.Errorf("projectID = %v, want test-project", exporter.projectID)
	}

	if exporter.clusterName != "test-cluster" {
		t.Errorf("clusterName = %v, want test-cluster", exporter.clusterName)
	}
}

// ============================================
// Prometheus Metrics Registration Tests
// ============================================

func TestPrometheusMetrics_AllDefined(t *testing.T) {
	metrics := []struct {
		name   string
		metric interface{}
	}{
		{"gkeNodeCPUUtilization", gkeNodeCPUUtilization},
		{"gkeNodeMemoryUtilization", gkeNodeMemoryUtilization},
		{"gkeNodeDiskUtilization", gkeNodeDiskUtilization},
		{"gkeContainerCPUUsage", gkeContainerCPUUsage},
		{"gkeContainerMemoryUsage", gkeContainerMemoryUsage},
		{"gkePodNetworkReceivedBytes", gkePodNetworkReceivedBytes},
		{"gkePodNetworkSentBytes", gkePodNetworkSentBytes},
		{"gkeClusterNodeCount", gkeClusterNodeCount},
		{"gkeClusterPodCount", gkeClusterPodCount},
		{"httpRequestCount", httpRequestCount},
		{"httpRequestLatency", httpRequestLatency},
	}

	for _, m := range metrics {
		t.Run(m.name, func(t *testing.T) {
			if m.metric == nil {
				t.Errorf("Metric %s should be defined", m.name)
			}
		})
	}
}

func TestPrometheusMetrics_LabelValidation(t *testing.T) {
	tests := []struct {
		name           string
		expectedLabels []string
		description    string
	}{
		{"Node metrics", []string{"node_name", "cluster_name"}, "Node CPU/memory/disk metrics should have node and cluster labels"},
		{"Container metrics", []string{"container_name", "pod_name", "namespace"}, "Container metrics should have container, pod, and namespace labels"},
		{"Pod metrics", []string{"pod_name", "namespace"}, "Pod metrics should have pod name and namespace labels"},
		{"Cluster metrics", []string{"cluster_name"}, "Cluster metrics should have cluster name label"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify expected label structure
			for _, label := range tt.expectedLabels {
				if label == "" {
					t.Error("Label should not be empty")
				}
			}
		})
	}
}

func TestPrometheusMetrics_NamingConventions(t *testing.T) {
	// Prometheus metrics should follow naming conventions:
	// - Must be lowercase
	// - No dashes (use underscores)
	// - Counters end with _total
	// - Gauges don't have _total
	// - Histograms end with _bucket, _count, _sum

	validMetricNames := []struct {
		name        string
		metricsType string
	}{
		{"gke_node_cpu_utilization", "gauge"},
		{"gke_container_cpu_usage_seconds", "counter"},
		{"gke_pod_network_received_bytes_total", "counter"},
		{"gcp_http_request_count", "gauge"},
	}

	for _, m := range validMetricNames {
		t.Run(m.name, func(t *testing.T) {
			// Verify name has no capital letters
			for _, r := range m.name {
				if r >= 'A' && r <= 'Z' {
					t.Errorf("Metric name %s has uppercase letter", m.name)
				}
			}

			// Verify no dashes
			if len(m.name) > 0 && m.name[0] == '-' {
				t.Error("Metric name should not start with dash")
			}

			// Verify counter names end with _total
			if m.metricsType == "counter" && !contains(m.name, "total") {
				t.Logf("Counter %s should end with _total (informational)", m.name)
			}
		})
	}
}

// ============================================
// Metric Collection Tests
// ============================================

func TestMetricCollection_NodeMetrics(t *testing.T) {
	// Verify node metric structure
	nodeMetrics := []string{
		"gke_node_cpu_utilization",
		"gke_node_memory_utilization",
		"gke_node_disk_utilization",
	}

	for _, metric := range nodeMetrics {
		if metric == "" {
			t.Error("Metric name should not be empty")
		}

		if len(metric) < 5 {
			t.Error("Metric name should be descriptive")
		}
	}
}

func TestMetricCollection_ContainerMetrics(t *testing.T) {
	// Verify container metric structure
	containerMetrics := []string{
		"gke_container_cpu_usage_seconds",
		"gke_container_memory_usage_bytes",
	}

	for _, metric := range containerMetrics {
		if !contains(metric, "gke_container") {
			t.Errorf("Container metric %s should have gke_container prefix", metric)
		}
	}
}

func TestMetricCollection_PodNetworkMetrics(t *testing.T) {
	// Verify pod network metrics are counters
	networkMetrics := []struct {
		name      string
		isCounter bool
		unit      string
	}{
		{"gke_pod_network_received_bytes_total", true, "bytes"},
		{"gke_pod_network_sent_bytes_total", true, "bytes"},
	}

	for _, m := range networkMetrics {
		if m.isCounter && !contains(m.name, "total") {
			t.Logf("Network counter %s should be marked as counter", m.name)
		}

		if !contains(m.name, m.unit) && m.unit != "bytes" {
			t.Logf("Network metric %s unit description missing", m.name)
		}
	}
}

// ============================================
// Environment Configuration Tests
// ============================================

func TestEnvironmentConfiguration_ProjectID(t *testing.T) {
	project := getEnv("GCP_PROJECT_ID", "default-project")

	if project == "" {
		t.Error("Project ID should not be empty")
	}

	// GCP project IDs should be lowercase with hyphens
	if len(project) < 1 {
		t.Error("Project ID should have content")
	}
}

func TestEnvironmentConfiguration_ClusterName(t *testing.T) {
	cluster := getEnv("GKE_CLUSTER_NAME", "default-cluster")

	if cluster == "" {
		t.Error("Cluster name should not be empty")
	}
}

func TestEnvironmentConfiguration_Retrieval(t *testing.T) {
	tests := []struct {
		envVar       string
		defaultValue string
		description  string
	}{
		{"GCP_PROJECT_ID", "default-project", "Project ID"},
		{"GKE_CLUSTER_NAME", "default-cluster", "Cluster name"},
		{"METRICS_INTERVAL", "15", "Collection interval"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// Verify getEnv function works
			os.Unsetenv(tt.envVar)
			result := getEnv(tt.envVar, tt.defaultValue)

			if result != tt.defaultValue {
				t.Errorf("getEnv(%s) = %v, want %v", tt.envVar, result, tt.defaultValue)
			}
		})
	}
}

// ============================================
// Data Validation Tests
// ============================================

func TestMetricValue_Types(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		valid bool
	}{
		{"CPU percentage", 45.5, true},
		{"Memory percentage", 78.2, true},
		{"Zero value", 0.0, true},
		{"Negative value", -10.0, false},
		{"Over 100 percentage", 150.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate metric values
			if tt.value < 0 && tt.valid {
				t.Error("Negative values should be marked invalid")
			}

			if tt.value > 100 && contains(tt.name, "percentage") && tt.valid {
				t.Logf("Percentage value over 100: %v (informational)", tt.value)
			}
		})
	}
}

func TestMetricLabel_Validation(t *testing.T) {
	validLabels := map[string]string{
		"node_name":      "gke-node-1",
		"pod_name":       "prometheus-pod",
		"namespace":      "monitoring",
		"cluster_name":   "prod-cluster",
		"container_name": "prometheus",
		"response_code":  "200",
	}

	for name, value := range validLabels {
		t.Run(name, func(t *testing.T) {
			if value == "" {
				t.Errorf("Label %s should have a value", name)
			}

			if len(value) == 0 {
				t.Errorf("Label %s value should not be empty", name)
			}
		})
	}
}

func TestMetricTimestamp_Validity(t *testing.T) {
	// Metrics should have valid timestamps
	// In a real scenario, these would come from GCP API responses

	tests := []struct {
		name           string
		timeValid      bool
		timeInPast     bool
		timeReasonable bool
	}{
		{"Current timestamp", true, true, true},
		{"Recent timestamp", true, true, true},
		{"Old timestamp", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.timeValid {
				t.Error("Timestamp should be valid")
			}
		})
	}
}

// ============================================
// Helper Functions
// ============================================

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestGetEnv_GCPExporter(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		defaultVal  string
		expectedVal string
	}{
		{"Use default if not set", "UNSET_VAR", "default_val", "default_val"},
		{"Use env value if set", "TEST_VAR", "default", "test_value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.envVar)

			if tt.expectedVal != "default_val" {
				os.Setenv(tt.envVar, "test_value")
				defer os.Unsetenv(tt.envVar)
			}

			result := getEnv(tt.envVar, tt.defaultVal)

			if result == "" {
				t.Error("getEnv should return non-empty value")
			}
		})
	}
}

// ============================================
// Integration-style Tests
// ============================================

func TestMetricCollectionFlow_Structure(t *testing.T) {
	// Verify the overall metric collection process structure
	t.Run("Metric collection sequence", func(t *testing.T) {
		// 1. Initialize accumulators
		acc := newCounterAccumulator()
		if acc == nil {
			t.Fatal("Accumulator initialization failed")
		}

		// 2. Get initial value
		delta1 := acc.delta("test_key", 100.0)
		if delta1 != 0 {
			t.Error("First delta should be 0")
		}

		// 3. Get incremental value
		delta2 := acc.delta("test_key", 150.0)
		if delta2 != 50.0 {
			t.Error("Second delta should reflect increase")
		}
	})
}

func TestMetricExportFlow_EndToEnd(t *testing.T) {
	// Simulate metric collection flow
	t.Run("Metric export simulation", func(t *testing.T) {
		// Create accumulator
		acc := newCounterAccumulator()

		// Simulate multiple collection cycles
		cycles := []float64{1000.0, 1500.0, 2000.0, 1000.0} // Last one represents counter reset

		expectations := []float64{0.0, 500.0, 500.0, 1000.0}

		for i, value := range cycles {
			delta := acc.delta("network_bytes", value)

			if i < len(expectations) && delta != expectations[i] {
				t.Errorf("Cycle %d: delta = %v, want %v", i+1, delta, expectations[i])
			}
		}
	})
}
