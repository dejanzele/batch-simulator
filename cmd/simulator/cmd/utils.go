package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pterm/pterm"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/k8s"
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
	printKWOKConfig()
	_ = pterm.
		DefaultBulletList.
		WithBulletStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightCyan)).
		WithItems([]pterm.BulletListItem{
			{Level: 1, Text: "node creator frequency = " + config.NodeCreatorFrequency.String()},
			{Level: 1, Text: "node creator requests  = " + fmt.Sprintf("%d", config.NodeCreatorRequests)},
			{Level: 1, Text: "node creator limit     = " + fmt.Sprintf("%d", config.NodeCreatorLimit)},
			{Level: 1, Text: "pod creator frequency  = " + config.PodCreatorFrequency.String()},
			{Level: 1, Text: "pod creator requests   = " + fmt.Sprintf("%d", config.PodCreatorRequests)},
			{Level: 1, Text: "pod creator limit      = " + fmt.Sprintf("%d", config.PodCreatorLimit)},
			{Level: 1, Text: "job creator frequency  = " + config.JobCreatorFrequency.String()},
			{Level: 1, Text: "job creator requests   = " + fmt.Sprintf("%d", config.JobCreatorRequests)},
			{Level: 1, Text: "job creator limit      = " + fmt.Sprintf("%d", config.JobCreatorLimit)},
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
func printMetricsEvery(ctx context.Context, interval time.Duration, manager *k8s.Manager, onFinished func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// clearScreen()
	oldNodeCreationMetrics, oldPodCreationMetrics, oldJobCreationMetrics := manager.Metrics()
	// Create a multi printer for managing multiple printers
	multi := pterm.DefaultMultiPrinter

	area, _ := pterm.
		DefaultArea.
		WithFullscreen(true).
		WithCenter(true).Start()
	defer func() { _ = area.Stop() }()
	printMetrics(area, oldNodeCreationMetrics, oldPodCreationMetrics, oldJobCreationMetrics)
	var nodeBar, podBar, jobBar *pterm.ProgressbarPrinter
	if config.NodeCreatorLimit > 0 {
		nodeBar, _ = pterm.
			DefaultProgressbar.
			WithWriter(multi.NewWriter()).
			WithTotal(config.NodeCreatorLimit).
			WithTitle("Node Creation Progress").
			Start()
	}
	if config.PodCreatorLimit > 0 {
		podBar, _ = pterm.
			DefaultProgressbar.
			WithWriter(multi.NewWriter()).
			WithTotal(config.PodCreatorLimit).
			WithTitle("Pod Creation Progress").
			Start()
	}
	if config.JobCreatorLimit > 0 {
		jobBar, _ = pterm.
			DefaultProgressbar.
			WithWriter(multi.NewWriter()).
			WithTotal(config.JobCreatorLimit).
			WithTitle("Job Creation Progress").
			Start()
	}

	_, _ = multi.Start()
	defer func() { _, _ = multi.Stop() }()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nodeCreationMetrics, podCreationMetrics, jobCreationMetrics := manager.Metrics()
			nodeCreationMetricsDelta, podCreationMetricsDelta, jobCreationMetricsDelta := calculateMetricsDelta(
				oldNodeCreationMetrics,
				nodeCreationMetrics,
				oldPodCreationMetrics,
				podCreationMetrics,
				oldJobCreationMetrics,
				jobCreationMetrics,
			)
			printMetrics(area, nodeCreationMetrics, podCreationMetrics, jobCreationMetrics)
			updateProgressBars(nodeBar, podBar, jobBar, nodeCreationMetricsDelta, podCreationMetricsDelta, jobCreationMetricsDelta)
			oldNodeCreationMetrics, oldPodCreationMetrics, oldJobCreationMetrics = nodeCreationMetrics, podCreationMetrics, jobCreationMetrics
			if finished(nodeCreationMetrics, podCreationMetrics, jobCreationMetrics) {
				if onFinished != nil {
					onFinished()
				}
				return
			}
		}
	}
}

// finished returns true if the node and pod creation metrics have reached their limits.
func finished(nodeCreationMetrics, podCreationMetrics, jobCreationMetrics ratelimiter.Metrics) bool {
	return nodeCreationMetrics.Executed == config.NodeCreatorLimit &&
		podCreationMetrics.Executed == config.PodCreatorLimit &&
		jobCreationMetrics.Executed == config.JobCreatorLimit
}

// calculateMetricsDelta calculates the delta between the old and new node and pod creation metrics.
func calculateMetricsDelta(
	oldNodeCreationMetrics, newNodeCreationMetrics, oldPodCreationMetrics, newPodCreationMetrics, oldJobCreationMetrics, newJobCreationMetrics ratelimiter.Metrics,
) (nodeCreationMetricsDelta, podCreationMetricsDelta, jobCreationMetricsDelta ratelimiter.Metrics) {
	nodeCreationMetricsDelta = calculateDelta(oldNodeCreationMetrics, newNodeCreationMetrics)
	podCreationMetricsDelta = calculateDelta(oldPodCreationMetrics, newPodCreationMetrics)
	jobCreationMetricsDelta = calculateDelta(oldJobCreationMetrics, newJobCreationMetrics)
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
func updateProgressBars(nodeBar, podBar, jobBar *pterm.ProgressbarPrinter, nodeMetrics, podMetrics, jobMetrics ratelimiter.Metrics) {
	if nodeBar != nil {
		nodeBar.Add(nodeMetrics.Succeeded + nodeMetrics.Failed)
	}
	if podBar != nil {
		podBar.Add(podMetrics.Succeeded + podMetrics.Failed)
	}
	if jobBar != nil {
		jobBar.Add(jobMetrics.Succeeded + jobMetrics.Failed)
	}
}

// printMetrics prints the node and pod creation metrics in a table.
func printMetrics(area *pterm.AreaPrinter, nodeMetrics, podMetrics, jobMetrics ratelimiter.Metrics) {
	data := pterm.TableData{
		{"Metric", "Node Creation", "Pod Creation", "Job Creation"},
		{"Executed", formatMetric(nodeMetrics.Executed), formatMetric(podMetrics.Executed), formatMetric(jobMetrics.Executed)},
		{"Failed", formatMetric(nodeMetrics.Failed), formatMetric(podMetrics.Failed), formatMetric(jobMetrics.Failed)},
		{"Succeeded", formatMetric(nodeMetrics.Succeeded), formatMetric(podMetrics.Succeeded), formatMetric(jobMetrics.Succeeded)},
	}

	table, _ := pterm.DefaultTable.WithHasHeader().WithData(data).Srender()
	area.Update(table)
}

// formatMetric formats the metric value to a string.
func formatMetric(metric int) string {
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
