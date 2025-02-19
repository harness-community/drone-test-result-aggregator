// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
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
	CompareBuildId      string `envconfig:"PLUGIN_COMPARE_BUILD_ID"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {

	logrus.Println("tool args.tool ", args.Tool)

	err := StoreResultsToInfluxDb(args)
	if err != nil {
		logrus.Println("error: ", err)
		return err
	}
	if args.CompareBuildResults || args.CompareBuildId != "" {
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
	var resultStr string
	var err error
	diffFileName := BuildResultsDiffCsv

	switch args.Tool {
	case JacocoTool:
		resultStr, err = CompareResults(JacocoTool, args)
	case JunitTool:
		resultStr, err = CompareJunitResults(JunitTool, args)
	case NunitTool:
		resultStr, err = CompareResults(NunitTool, args)
	case TestNgTool:
		resultStr, err = CompareResults(TestNgTool, args)
	default:
		errStr := fmt.Sprintf("Tool type %s not supported to compare builds", args.Tool)
		return errors.New(errStr)
	}

	if err != nil {
		logrus.Println("Unable to compare results ", err)
		return err
	}
	err = ExportComparisonResults(diffFileName, resultStr, TestResultsDiffFileOutputVar)
	if err != nil {
		logrus.Println("Unable to export comparison results ", err)
		return err
	}
	return nil
}

func ExportComparisonResults(resultFileName, resultStr, outputVarName string) error {
	err := WriteStrToFile(resultFileName, resultStr)
	if err != nil {
		logrus.Println("Unable to write comparison results to file ", err)
	}
	err = WriteToEnvVariable(outputVarName, resultFileName)
	if err != nil {
		logrus.Println("Unable to write comparison results to env variable ", err)
	}
	return err
}
