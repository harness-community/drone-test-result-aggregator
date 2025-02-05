// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// plugin params
	Tool                string `envconfig:"PLUGIN_TOOL"`
	Command             string `envconfig:"PLUGIN_COMMAND"`
	ReportsDir          string `envconfig:"PLUGIN_REPORTS_DIR"`
	ReportsName         string `envconfig:"PLUGIN_REPORTS_NAME"`
	IncludePattern      string `envconfig:"PLUGIN_INCLUDE_PATTERN"`
	DbUrl               string `envconfig:"PLUGIN_INFLUXDB_URL"`
	DbToken             string `envconfig:"PLUGIN_INFLUXDB_TOKEN"`
	DbOrg               string `envconfig:"PLUGIN_INFLUXDB_ORG"`
	DbBucket            string `envconfig:"PLUGIN_INFLUXDB_BUCKET"`
	GroupName           string `envconfig:"PLUGIN_GROUP"`
	CompareBuildResults bool   `envconfig:"PLUGIN_COMPARE_BUILD_RESULTS"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {

	logrus.Println("tool args.tool ", args.Tool)

	err := StoreResultsToInfluxDb(args)
	if err != nil {
		logrus.Println("error: ", err)
		return err
	}
	if args.CompareBuildResults {
		err = CompareBuildResults(args)
		if err != nil {
			logrus.Println("error: ", err)
			return err
		}
	}
	return nil
}

func StoreResultsToInfluxDb(args Args) error {
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
	errStr := fmt.Sprintf("Tool type %s not supported to aggregate", args.Tool)
	return errors.New(errStr)
}

func CompareBuildResults(args Args) error {

	var retMap map[string]interface{}

	fmt.Println("Tool: ", args.Tool)
	var err error
	switch args.Tool {
	case JacocoTool:
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>> jacoco ")
		retMap, err = CompareResults(JacocoTool, args)
	default:
		errStr := fmt.Sprintf("Tool type %s not supported to compare builds", args.Tool)
		return errors.New(errStr)
	}
	// convert map to json
	jsonBytes, err := json.Marshal(retMap)
	if err != nil {
		logrus.Println("Error converting map to json: ", err)
		return err
	}

	fmt.Println("UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU")
	fmt.Println(string(jsonBytes))
	return err
}
