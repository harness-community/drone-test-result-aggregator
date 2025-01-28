// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"github.com/sirupsen/logrus"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// plugin params
	Tool           string `envconfig:"PLUGIN_TOOL"`
	Command        string `envconfig:"PLUGIN_COMMAND"`
	ReportsDir     string `envconfig:"PLUGIN_REPORTS_DIR"`
	ReportsName    string `envconfig:"PLUGIN_REPORTS_NAME"`
	IncludePattern string `envconfig:"PLUGIN_INCLUDE_PATTERN"`
	DbUrl          string `envconfig:"PLUGIN_DB_URL"`
	DbToken        string `envconfig:"PLUGIN_DB_TOKEN"`
	DbOrg          string `envconfig:"PLUGIN_DB_ORG"`
	DbBucket       string `envconfig:"PLUGIN_DB_BUCKET"`
	GroupName      string `envconfig:"PLUGIN_GROUP"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	logrus.Println("tool args.tool ", args.Tool)

	if args.Command == SaveToDb {
		switch args.Tool {
		case JacocoTool:
			aggregator := GetNewJacocoAggregator(args.ReportsDir, args.ReportsName, args.IncludePattern,
				args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket)
			return aggregator.Aggregate(args.GroupName)
		case JunitTool:
			aggregator := GetNewJunitAggregator(args.ReportsDir, args.ReportsName, args.IncludePattern,
				args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket)
			return aggregator.Aggregate(args.GroupName)
		case NunitTool:
			aggregator := GetNewNunitAggregator(args.ReportsDir, args.ReportsName, args.IncludePattern,
				args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket)
			return aggregator.Aggregate(args.GroupName)
		case TestNgTool:
			aggregator := GetNewTestNgAggregator(args.ReportsDir, args.ReportsName, args.IncludePattern,
				args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket)
			return aggregator.Aggregate(args.GroupName)
		}
	}
	return nil
}
