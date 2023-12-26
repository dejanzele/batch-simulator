package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pterm/pterm"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/kubernetes"
	"github.com/dejanzele/batch-simulator/internal/ratelimiter"
)

// printKWOKConfig prints the configuration for k8s and kwok
func printKWOKConfig() {
	printConfigSection()
	_ = pterm.
		DefaultBulletList.
		WithBulletStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithItems([]pterm.BulletListItem{
			{Level: 1, Text: "kwok namespace = " + config.KWOKNamespace},
			{Level: 1, Text: "kubeconfig     = " + config.Kubeconfig},
			{Level: 1, Text: "qps            = " + fmt.Sprintf("%f", config.QPS)},
			{Level: 1, Text: "burst          = " + fmt.Sprintf("%d", config.Burst)},
		}).Render()
}

// printSimulationConfig prints the configuration for the simulation.
func printSimulationConfig() {
	printConfigSection()
	_ = pterm.
		DefaultBulletList.
		WithBulletStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithItems([]pterm.BulletListItem{
			{Level: 1, Text: "node creator frequency = " + fmt.Sprintf("%f", config.NodeCreatorFrequency.Seconds())},
			{Level: 1, Text: "node creator requests  = " + fmt.Sprintf("%d", config.NodeCreatorRequests)},
			{Level: 1, Text: "node creator limit     = " + fmt.Sprintf("%d", config.NodeCreatorLimit)},
			{Level: 1, Text: "pod creator frequency  = " + fmt.Sprintf("%f", config.PodCreatorFrequency.Seconds())},
			{Level: 1, Text: "pod creator requests   = " + fmt.Sprintf("%d", config.PodCreatorRequests)},
			{Level: 1, Text: "pod creator limit      = " + fmt.Sprintf("%d", config.PodCreatorLimit)},
		}).Render()
}

// printConfigSection prints the configuration section.
func printConfigSection() {
	pterm.DefaultSection.Println("config")
}

// printMetricsEvery prints the metrics every interval.
// - ctx is the context that should be used for the ticker.
// - interval is the interval at which the metrics should be printed.
// - manager is the kubernetes manager which can provide metrics.
func printMetricsEvery(ctx context.Context, interval time.Duration, manager *kubernetes.Manager, onFinished func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// clearScreen()
	oldNodeCreationMetrics, oldPodCreationMetrics := manager.Metrics()
	// Create a multi printer for managing multiple printers
	multi := pterm.DefaultMultiPrinter

	area, _ := pterm.
		DefaultArea.
		WithFullscreen(true).
		WithCenter(true).Start()
	defer func() { _ = area.Stop() }()
	printMetrics(area, oldNodeCreationMetrics, oldPodCreationMetrics)
	nodeBar, _ := pterm.
		DefaultProgressbar.
		WithWriter(multi.NewWriter()).
		WithTotal(int(config.NodeCreatorLimit)).
		WithTitle("Node Creation Progress").
		Start()
	podBar, _ := pterm.
		DefaultProgressbar.
		WithWriter(multi.NewWriter()).
		WithTotal(int(config.PodCreatorLimit)).
		WithTitle("Pod Creation Progress").
		Start()
	_, _ = multi.Start()
	defer func() { _, _ = multi.Stop() }()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nodeCreationMetrics, podCreationMetrics := manager.Metrics()
			nodeCreationMetricsDelta, podCreationMetricsDelta := calculateMetricsDelta(
				oldNodeCreationMetrics,
				nodeCreationMetrics,
				oldPodCreationMetrics,
				podCreationMetrics,
			)
			printMetrics(area, nodeCreationMetrics, podCreationMetrics)
			updateProgressBars(nodeBar, podBar, nodeCreationMetricsDelta, podCreationMetricsDelta)
			oldNodeCreationMetrics, oldPodCreationMetrics = nodeCreationMetrics, podCreationMetrics
			if finished(nodeCreationMetrics, podCreationMetrics) {
				if onFinished != nil {
					onFinished()
				}
				return
			}
		}
	}
}

// finished returns true if the node and pod creation metrics have reached their limits.
func finished(nodeCreationM, podCreationM ratelimiter.Metrics) bool {
	return nodeCreationM.Executed == config.NodeCreatorLimit && podCreationM.Executed == config.PodCreatorLimit
}

// calculateMetricsDelta calculates the delta between the old and new node and pod creation metrics.
func calculateMetricsDelta(
	oldNodeCreationMetrics, newNodeCreationMetrics, oldPodCreationMetrics, newPodCreationMetrics ratelimiter.Metrics,
) (nodeCreationMetricsDelta, podCreationMetricsDelta ratelimiter.Metrics) {
	nodeCreationMetricsDelta = calculateDelta(oldNodeCreationMetrics, newNodeCreationMetrics)
	podCreationMetricsDelta = calculateDelta(oldPodCreationMetrics, newPodCreationMetrics)
	return
}

// calculateDelta calculates the delta between the old and new metric.
func calculateDelta(previous, latest ratelimiter.Metrics) ratelimiter.Metrics {
	return ratelimiter.Metrics{
		Executed:  latest.Executed - previous.Executed,
		Failed:    latest.Failed - previous.Failed,
		Succeeded: latest.Succeeded - previous.Succeeded,
	}
}

// updateProgressBars updates the progress bars with node and pod creation metrics.
func updateProgressBars(nodeBar, podBar *pterm.ProgressbarPrinter, nodeMetrics, podMetrics ratelimiter.Metrics) {
	nodeBar.Add(int(nodeMetrics.Succeeded + nodeMetrics.Failed))
	podBar.Add(int(podMetrics.Succeeded + podMetrics.Failed))
}

// printMetrics prints the node and pod creation metrics in a table.
func printMetrics(area *pterm.AreaPrinter, nodeMetrics, podMetrics ratelimiter.Metrics) {
	data := pterm.TableData{
		{"Metric", "Node Creation", "Pod Creation"},
		{"Executed", formatMetric(nodeMetrics.Executed), formatMetric(podMetrics.Executed)},
		{"Failed", formatMetric(nodeMetrics.Failed), formatMetric(podMetrics.Failed)},
		{"Succeeded", formatMetric(nodeMetrics.Succeeded), formatMetric(podMetrics.Succeeded)},
	}

	table, _ := pterm.DefaultTable.WithHasHeader().WithData(data).Srender()
	area.Update(table)
}

// formatMetric formats the metric value to a string.
func formatMetric(metric int32) string {
	return pterm.Sprintf("%d", metric)
}

// blip is a helper function which is used to slow down the output.
func blip() {
	time.Sleep(200 * time.Millisecond)
}

func getLogLevel() slog.Level {
	switch {
	case config.Silent:
		return slog.Level(10)
	case config.Debug:
		return slog.LevelDebug
	case config.Verbose:
		return slog.LevelInfo
	default:
		return slog.LevelWarn
	}
}
