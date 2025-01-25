// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"harness-community/drone-test-result-aggregator/plugin/jacoco"
	"harness-community/drone-test-result-aggregator/plugin/utils"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// plugin params
	Tool           string `envconfig:"PLUGIN_TOOL"`
	ReportsDir     string `envconfig:"PLUGIN_REPORTS_DIR"`
	ReportsName    string `envconfig:"PLUGIN_REPORTS_NAME"`
	IncludePattern string `envconfig:"PLUGIN_INCLUDE_PATTERN"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	fmt.Println("tool args.tool ", args.Tool)

	switch args.Tool {
	case utils.JacocoTool:
		aggregator := jacoco.JacocoAggregator{
			ReportsDir:  args.ReportsDir,
			ReportsName: args.ReportsName,
			Includes:    args.IncludePattern,
		}
		return aggregator.Aggregate()
	}
	return nil
}
