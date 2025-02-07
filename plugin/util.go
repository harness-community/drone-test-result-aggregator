// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sirupsen/logrus"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	JacocoTool        = "jacoco"
	JunitTool         = "junit"
	NunitTool         = "nunit"
	TestNgTool        = "testng"
	SaveToDb          = "save-to-db"
	PipeLineIdEnvVar  = "HARNESS_PIPELINE_ID"
	BuildNumberEnvVar = "HARNESS_BUILD_ID"
)

type ResultBasicInfo struct {
	PipelineId string
	BuildId    string
	Name       string
	Type       string
	Status     string
}

type DbCredentials struct {
	InfluxDBURL   string
	InfluxDBToken string
	Organization  string
	Bucket        string
}

type BuildResultCompare struct {
	Tool               string
	CurrentPipelineId  string
	CurrentBuildId     string
	PreviousPipelineId string
	PreviousBuildId    string
}

type ResultDiff struct {
	FieldName            string
	CurrentBuildValue    float64
	PreviousBuildValue   float64
	Difference           float64
	PercentageDifference float64
	IsCompareValid       bool
}

func GetNewBuildResultCompare(tool, currentPipelineId, currentBuildId, previousPipelineId, previousBuildId string) BuildResultCompare {
	return BuildResultCompare{
		Tool:               tool,
		CurrentPipelineId:  currentPipelineId,
		CurrentBuildId:     currentBuildId,
		PreviousPipelineId: previousPipelineId,
		PreviousBuildId:    previousBuildId,
	}
}

func GetXmlReportData[T any](reportsRootDir string, patterns []string) ([]T, error) {

	logrus.Println("GetXmlReportData: reportsRootDir ==  ", reportsRootDir)

	var xmlReportFiles []string
	var xmlFileReportDataList []T

	for _, pattern := range patterns {
		logrus.Println("junit: pattern ==  ", pattern)
		logrus.Println("junit: reportsRootDir ==  ", reportsRootDir)
		tmpReportDir := os.DirFS(reportsRootDir)
		relPattern := strings.TrimPrefix(pattern, reportsRootDir+"/")
		filesList, err := doublestar.Glob(tmpReportDir, relPattern)
		if err != nil {
			logrus.Println("Include patterns not found ", err.Error())
			return xmlFileReportDataList, err
		}
		xmlReportFiles = append(xmlReportFiles, filesList...)
	}

	logrus.Println("xmlReportFiles ==  ", xmlReportFiles)
	logrus.Println("len(xmlReportFiles) ", len(xmlReportFiles))

	for _, xmlReportFile := range xmlReportFiles {
		logrus.Println("Processing junit result file: ", xmlReportFile)
		tmpXmlReportFile := filepath.Join(reportsRootDir, xmlReportFile)
		report := ParseXmlReport[T](tmpXmlReportFile)
		reportBytes, err := json.Marshal(report)
		if err != nil {
			logrus.Printf("Error marshalling report: %v", err)
		}

		xmlFileReport, err := ToStructFromJsonString[T](string(reportBytes))
		if err != nil {
			logrus.Printf("Error converting json to struct: %v", err)
			return xmlFileReportDataList, err
		}

		xmlFileReportDataList = append(xmlFileReportDataList, xmlFileReport)
	}

	return xmlFileReportDataList, nil
}

func ParseXmlReport[T any](filename string) T {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Println("Error opening XML file: ", err)
		logrus.Fatalf("Error opening XML file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		logrus.Println("Error reading XML file ", err)
	}

	var report T
	err = xml.Unmarshal(data, &report)
	if err != nil {
		logrus.Printf("Error unmarshalling XML: %v", err)
	}
	return report
}

func Aggregate[T any](reportsDir, includes string,
	dbUrl, dbToken, dbOrg, dbBucket, measurementName, groupName string,
	calculateAggregate func(testNgAggregatorList []T) T,
	getDataMaps func(pipelineId, buildNumber string, aggregateData T) (map[string]string, map[string]interface{}),
	showBuildStats func(tagsMap map[string]string, fieldsMap map[string]interface{}) error) error {

	reportsRootDir := reportsDir
	patterns := strings.Split(includes, ",")

	aggregatorList, err := GetXmlReportData[T](reportsRootDir, patterns)
	if err != nil {
		logrus.Println("Error getting xml report data: ", err.Error())
		return err
	}

	totalAggregate := calculateAggregate(aggregatorList)
	logrus.Println("Total Aggregate: ", totalAggregate)

	pipelineId, buildNumber, err := GetPipelineInfo()
	if err != nil {
		logrus.Println("Error getting pipeline info: ", err.Error())
		return err
	}

	tagsMap, fieldsMap := getDataMaps(pipelineId, buildNumber, totalAggregate)
	err = showBuildStats(tagsMap, fieldsMap)
	if err != nil {
		logrus.Println("Error showing build stats: ", err.Error())
		return err
	}
	err = PersistToInfluxDb(dbUrl, dbToken, dbOrg, dbBucket, measurementName, groupName, tagsMap, fieldsMap)

	return err
}

func PersistToInfluxDb(dbUrl, dbToken, dbOrganisation, dbBucket, measurementName, groupName string,
	tagsMap map[string]string, fieldsMap map[string]interface{}) error {

	tagsMap["group"] = groupName
	client := influxdb2.NewClient(dbUrl, dbToken)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(dbOrganisation, dbBucket)

	point := influxdb2.NewPoint(
		measurementName,
		tagsMap,
		fieldsMap,
		time.Now())
	err := writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		logrus.Println("Error writing point: ", err)
		return err
	}
	logrus.Println("Data persisted successfully to InfluxDB.")
	return nil
}

func GetPipelineInfo() (string, string, error) {
	pipelineId := os.Getenv(PipeLineIdEnvVar)
	buildNumber := os.Getenv(BuildNumberEnvVar)

	if pipelineId == "" || buildNumber == "" {
		return "", "", fmt.Errorf("PipelineId or BuildNumber not found in the environment")
	}

	return pipelineId, buildNumber, nil
}

func GetPreviousBuildId(measurementName, influxURL, token, org, bucket, currentPipelineId, groupId, currentBuildId string) (prevBuildId int, err error) {
	client := influxdb2.NewClient(influxURL, token)
	defer client.Close()

	query := fmt.Sprintf(`
	from(bucket: "%s")
	  |> range(start: -1y)
	  |> filter(fn: (r) => r._measurement == "%s")
	  |> filter(fn: (r) => r.pipelineId == "%s")
	  |> filter(fn: (r) => r.group == "%s")
	  |> keep(columns: ["buildId"])
	  |> distinct(column: "buildId")
	  |> sort(columns: ["buildId"], desc: true)
	`, bucket, measurementName, currentPipelineId, groupId)

	queryAPI := client.QueryAPI(org)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, fmt.Errorf("failed to query InfluxDB: %w", err)
	}

	var buildIds []int
	for result.Next() {
		buildIdStr := result.Record().ValueByKey("buildId")
		if buildIdStr == nil {
			continue
		}
		buildId, convErr := strconv.Atoi(buildIdStr.(string))
		if convErr != nil {
			continue
		}
		buildIds = append(buildIds, buildId)
	}

	if len(buildIds) == 0 {
		fmt.Println("No build IDs found for measurement=", measurementName, "pipelineId=", currentPipelineId, "group=", groupId)
		return 0, fmt.Errorf("no build IDs found for measurement=%s, pipelineId=%s, group=%s",
			measurementName, currentPipelineId, groupId)
	}

	currentBuild, err := strconv.Atoi(currentBuildId)
	if err != nil {
		fmt.Println("Invalid currentBuildId: ", currentBuildId)
		return 0, fmt.Errorf("invalid currentBuildId: %s", currentBuildId)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(buildIds)))
	for _, id := range buildIds {
		if id < currentBuild {
			return id, nil
		}
	}

	return 0, fmt.Errorf("no previous build ID found for %d", currentBuild)
}

func CompareResults(tool string, args Args) (string, error) {
	var resultStr string
	currentPipelineId, currentBuildNumber, err := GetPipelineInfo()
	if err != nil {
		fmt.Println("CompareResults Error getting pipeline info: ", err)
		return resultStr, err
	}

	previousBuildId, err := GetPreviousBuildId(tool, args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket, currentPipelineId, args.GroupName, currentBuildNumber)
	if err != nil {
		fmt.Println("CompareResults Error getting previous build id: ", err)
		return resultStr, err
	}

	resultStr, err = GetComparedDifferences(tool, args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket, currentPipelineId, args.GroupName, currentBuildNumber, strconv.Itoa(previousBuildId))
	if err != nil {
		fmt.Println("CompareResults Error getting compared differences: ", err)
		return resultStr, err
	}
	return resultStr, nil
}

func GetComparedDifferences(measurementName, influxURL, token, org, bucket, currentPipelineId, groupId, currentBuildId, previousBuildId string) (string, error) {
	client := influxdb2.NewClient(influxURL, token)
	defer client.Close()

	currentValues, err := GetStoredBuildResults(client, org, bucket, measurementName, currentPipelineId, groupId, currentBuildId)
	if err != nil {
		fmt.Println("GetComparedDifferences Error fetching current build values: ", err)
		return "", fmt.Errorf("error fetching current build values: %w", err)
	}

	previousValues, err := GetStoredBuildResults(client, org, bucket, measurementName, currentPipelineId, groupId, previousBuildId)
	if err != nil {
		fmt.Println("GetComparedDifferences Error fetching previous build values: ", err)
		return "", fmt.Errorf("error fetching previous build values: %w", err)
	}

	return ComputeBuildResultDifferences(currentValues, previousValues), nil
}

func GetStoredBuildResults(client influxdb2.Client, org, bucket, measurementName,
	pipelineId, groupId, buildId string) (map[string]float64, error) {

	query := fmt.Sprintf(`
	from(bucket: "%s")
	  |> range(start: -1y)
	  |> filter(fn: (r) => r._measurement == "%s")
	  |> filter(fn: (r) => r.pipelineId == "%s")
	  |> filter(fn: (r) => r.group == "%s")
	  |> filter(fn: (r) => r.buildId == "%s")
	  |> keep(columns: ["_field", "_value"])
	`, bucket, measurementName, pipelineId, groupId, buildId)

	queryAPI := client.QueryAPI(org)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		fmt.Println("fetchBuildFieldValues Error querying InfluxDB: ", err)
		return nil, fmt.Errorf("failed to query InfluxDB: %w", err)
	}

	fieldValues := make(map[string]float64)
	for result.Next() {
		fieldName := result.Record().ValueByKey("_field")
		value := result.Record().ValueByKey("_value")

		if fieldName == nil || value == nil {
			continue
		}

		fieldNameStr := fieldName.(string)
		valueFloat, convErr := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
		if convErr != nil {
			continue
		}

		fieldValues[fieldNameStr] = valueFloat
	}

	return fieldValues, nil
}

func ComputeBuildResultDifferences(currentValues, previousValues map[string]float64) string {
	resultDiffs := []ResultDiff{}
	allFields := make(map[string]struct{})

	for field := range currentValues {
		allFields[field] = struct{}{}
	}
	for field := range previousValues {
		allFields[field] = struct{}{}
	}

	for field := range allFields {
		currentValue, currentExists := currentValues[field]
		previousValue, previousExists := previousValues[field]

		isCompareValid := currentExists && previousExists
		if !currentExists {
			currentValue = 0
		}
		if !previousExists {
			previousValue = 0
		}

		diff := currentValue - previousValue
		percentageDiff := 0.0
		if previousValue != 0 {
			percentageDiff = (diff / math.Abs(previousValue)) * 100
		}

		resultDiffs = append(resultDiffs, ResultDiff{
			FieldName:            field,
			CurrentBuildValue:    currentValue,
			PreviousBuildValue:   previousValue,
			Difference:           diff,
			PercentageDifference: percentageDiff,
			IsCompareValid:       isCompareValid,
		})
	}

	resultFormat := "ResultType, CurrentBuild, PreviousBuild, Difference, PercentageDifference " + "\n"
	var resultsStr string
	resultsStr = resultsStr + resultFormat

	for _, diff := range resultDiffs {
		if diff.IsCompareValid {
			row := fmt.Sprintf("%s, %.2f, %.2f, %.2f, %.2f",
				diff.FieldName, diff.CurrentBuildValue, diff.PreviousBuildValue, diff.Difference, diff.PercentageDifference) + "\n"
			resultsStr = resultsStr + row
		}
	}
	fmt.Println("")
	fmt.Println("Comparison results with previous build:")
	ShowDiffAsTable(currentValues, previousValues)
	fmt.Println("")
	return resultsStr
}

func ShowDiffAsTable(currentValues, previousValues map[string]float64) {
	// Collect all unique fields from both maps
	allFields := make(map[string]struct{})
	for key := range currentValues {
		allFields[key] = struct{}{}
	}
	for key := range previousValues {
		allFields[key] = struct{}{}
	}

	// Sort fields alphabetically for consistent output
	var sortedFields []string
	for field := range allFields {
		sortedFields = append(sortedFields, field)
	}
	sort.Strings(sortedFields)

	// Print table header with borders
	fmt.Println("+------------------+---------------+---------------+------------+------------------------+")
	fmt.Println("|   Result Type    | Current Build | Previous Build | Difference | Percentage Difference |")
	fmt.Println("+------------------+---------------+---------------+------------+------------------------+")

	// Print each row
	for _, field := range sortedFields {
		currentVal := currentValues[field]
		previousVal := previousValues[field]

		diff := currentVal - previousVal
		percentageDiff := 0.0
		if previousVal != 0 {
			percentageDiff = (diff / math.Abs(previousVal)) * 100
		}

		// Print formatted row
		fmt.Printf("| %-16s | %-13d | %-13d | %-10d | %-22.2f%% |\n",
			field, int(currentVal), int(previousVal), int(diff), percentageDiff)
	}

	// Print table footer
	fmt.Println("+------------------+---------------+---------------+------------+------------------------+")
}

func ShowDiffAsTable2(currentValues, previousValues map[string]float64) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(writer, "Result Type\tCurrent Build\tPrevious Build\tDifference\tPercentage Difference")
	allFields := make(map[string]struct{})
	for key := range currentValues {
		allFields[key] = struct{}{}
	}
	for key := range previousValues {
		allFields[key] = struct{}{}
	}
	for field := range allFields {
		currentVal := currentValues[field]
		previousVal := previousValues[field]

		diff := currentVal - previousVal
		percentageDiff := 0.0
		if previousVal != 0 {
			percentageDiff = (diff / math.Abs(previousVal)) * 100
		}
		fmt.Fprintf(writer, "%s\t%d\t%d\t%d\t%.2f%%\n", field, int(currentVal), int(previousVal), int(diff), percentageDiff)
	}
	writer.Flush()
}

func WriteToEnvVariable(key string, value interface{}) error {

	outputFile, err := os.OpenFile(os.Getenv("DRONE_OUTPUT"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer outputFile.Close()

	valueStr := fmt.Sprintf("%v", value)

	_, err = fmt.Fprintf(outputFile, "%s=%s\n", key, valueStr)
	if err != nil {
		return fmt.Errorf("failed to write to env: %w", err)
	}

	return nil
}

func ToStructFromJsonString[T any](jsonStr string) (T, error) {
	var v T
	err := json.Unmarshal([]byte(jsonStr), &v)
	return v, err
}

func ToJsonStringFromStruct[T any](v T) (string, error) {
	jsonBytes, err := json.Marshal(v)

	if err == nil {
		return string(jsonBytes), nil
	}

	return "", err
}
