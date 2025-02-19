// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"encoding/csv"
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
	"time"
)

const (
	JacocoTool                   = "jacoco"
	JunitTool                    = "junit"
	NunitTool                    = "nunit"
	TestNgTool                   = "testng"
	PipeLineIdEnvVar             = "HARNESS_PIPELINE_ID"
	BuildNumberEnvVar            = "HARNESS_BUILD_ID"
	TestResultsDiffFileOutputVar = "TEST_RESULTS_DIFF_FILE"
	BuildResultsDiffCsv          = "build_results_diff.csv"
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
	FieldName            string  `json:"type"`
	CurrentBuildValue    float64 `json:"current_build"`
	PreviousBuildValue   float64 `json:"previous_build"`
	Difference           float64 `json:"difference"`
	PercentageDifference float64 `json:"percentage_difference"`
	IsCompareValid       bool    `json:"-"`
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
		tmpReportDir := os.DirFS(reportsRootDir)
		relPattern := strings.TrimPrefix(pattern, reportsRootDir+"/")
		filesList, err := doublestar.Glob(tmpReportDir, relPattern)
		if err != nil {
			logrus.Println("Include patterns not found ", err.Error())
			return xmlFileReportDataList, err
		}
		xmlReportFiles = append(xmlReportFiles, filesList...)
	}

	for _, xmlReportFile := range xmlReportFiles {
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
	getDataMaps func(pipelineId,
		buildNumber string, aggregateData T) (map[string]string, map[string]interface{}),
	showBuildStats func(tagsMap map[string]string,
		fieldsMap map[string]interface{}) error) (map[string]string, map[string]interface{}, error) {

	tagsMap := map[string]string{}
	fieldsMap := map[string]interface{}{}

	reportsRootDir := reportsDir
	patterns := strings.Split(includes, ",")

	aggregatorList, err := GetXmlReportData[T](reportsRootDir, patterns)
	if err != nil {
		logrus.Println("Error getting xml report data: ", err.Error())
		return tagsMap, fieldsMap, err
	}

	totalAggregate := calculateAggregate(aggregatorList)
	logrus.Println("Total Aggregate: ", totalAggregate)

	pipelineId, buildNumber, err := GetPipelineInfo()
	if err != nil {
		logrus.Println("Error getting pipeline info: ", err.Error())
		return tagsMap, fieldsMap, err
	}

	tagsMap, fieldsMap = getDataMaps(pipelineId, buildNumber, totalAggregate)
	err = showBuildStats(tagsMap, fieldsMap)
	if err != nil {
		logrus.Println("Error showing build stats: ", err.Error())
		return tagsMap, fieldsMap, err
	}

	if dbUrl != "" && dbToken != "" && dbOrg != "" && dbBucket != "" {
		err = PersistToInfluxDb(dbUrl, dbToken, dbOrg, dbBucket, measurementName, groupName, tagsMap, fieldsMap)
		if err != nil {
			logrus.Println("Error persisting data to InfluxDB: ", err.Error())
			return tagsMap, fieldsMap, err
		}
	}

	return tagsMap, fieldsMap, err
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

func GetPreviousBuildId(measurementName, influxURL, token, org, bucket,
	currentPipelineId, groupId, currentBuildId string, args Args) (int, error) {

	if args.CompareBuildId != "" {
		prevBuildId, convErr := strconv.Atoi(args.CompareBuildId)
		if convErr != nil {
			logrus.Println("Error converting previous build ID ", args.CompareBuildId, " to int: ", convErr)
			return 0, fmt.Errorf("error converting previous build ID to int: %w", convErr)
		}
		fmt.Println("Comparing against build ID: ", prevBuildId)
		return prevBuildId, nil
	}

	client := influxdb2.NewClient(influxURL, token)
	defer client.Close()

	query := fmt.Sprintf(`
	from(bucket: "%s")
	  |> range(start: -1y)
	  |> filter(fn: (r) => r._measurement == "%s" and r.pipelineId == "%s" and r.group == "%s")
	  |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
	  |> sort(columns: ["buildId"], desc: true)
	`, bucket, measurementName, currentPipelineId, groupId)

	queryAPI := client.QueryAPI(org)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		logrus.Println("Error querying InfluxDB: ", err)
		return 0, fmt.Errorf("failed to query InfluxDB: %w", err)
	}

	currentBuild, err := strconv.Atoi(currentBuildId)
	if err != nil {
		logrus.Println("Invalid currentBuildId: ", currentBuildId)
		return 0, fmt.Errorf("invalid currentBuildId: %s", currentBuildId)
	}

	var prevBuildId = 0
	var buildId int
	found := false
	var convErr error

	for result.Next() {
		buildIdStr := result.Record().ValueByKey("buildId")
		if buildIdStr == nil {
			continue
		}

		prevBuildId = buildId
		buildId, convErr = strconv.Atoi(buildIdStr.(string))
		if convErr != nil {
			logrus.Println("Error converting buildId ", buildIdStr, "to int error: ", convErr)
			continue
		}

		if buildId >= currentBuild {
			found = true
			break
		}
	}

	if !found {
		logrus.Println("No previous build ID found for ", currentBuild)
		return 0, fmt.Errorf("no previous build ID found for %d", currentBuild)
	}

	fmt.Println("Previous build ID found: ", prevBuildId)
	return prevBuildId, nil
}

func CompareResults(tool string, args Args) (string, error) {
	var resultStr string
	currentPipelineId, currentBuildNumber, err := GetPipelineInfo()
	if err != nil {
		fmt.Println("CompareResults Error getting pipeline info: ", err)
		return resultStr, err
	}

	previousBuildId, err := GetPreviousBuildId(tool, args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket, currentPipelineId, args.GroupName, currentBuildNumber, args)
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

	diffStr, err := ComputeBuildResultDifferences(currentValues, previousValues)
	if err != nil {
		fmt.Println("GetComparedDifferences Error computing differences: ", err)
		return "", err
	}
	return diffStr, nil
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

func ComputeBuildResultDifferences(currentValues, previousValues map[string]float64) (string, error) {
	allFields := make(map[string]struct{})

	for field := range currentValues {
		allFields[field] = struct{}{}
	}
	for field := range previousValues {
		allFields[field] = struct{}{}
	}

	var csvBuffer strings.Builder
	writer := csv.NewWriter(&csvBuffer)

	err := writer.Write([]string{"Field Name", "Current", "Previous", "Difference", "Percentage Difference"})
	if err != nil {
		logrus.Println("Error writing to CSV writer: ", err)
		return "", err
	}

	for field := range allFields {
		currentValue, currentExists := currentValues[field]
		previousValue, previousExists := previousValues[field]

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

		err = writer.Write([]string{
			field,
			fmt.Sprintf("%.2f", currentValue),
			fmt.Sprintf("%.2f", previousValue),
			fmt.Sprintf("%.2f", diff),
			fmt.Sprintf("%.2f%%", percentageDiff),
		})
		if err != nil {
			logrus.Println("Error writing to CSV writer: ", err)
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Println("Error flushing CSV writer:", err)
		return "", err
	}

	fmt.Println("")
	fmt.Println("Comparison results with previous build:")
	ShowDiffAsTable(currentValues, previousValues)
	fmt.Println("")

	return csvBuffer.String(), nil
}

func ShowDiffAsTable(currentValues, previousValues map[string]float64) {
	allFields := make(map[string]struct{})
	for key := range currentValues {
		allFields[key] = struct{}{}
	}
	for key := range previousValues {
		allFields[key] = struct{}{}
	}

	var sortedFields []string
	for field := range allFields {
		sortedFields = append(sortedFields, field)
	}
	sort.Strings(sortedFields)

	maxFieldLen := len("Result Type")
	maxValueLen := len("Current Build")
	maxDiffLen := len("Difference")
	maxPercentDiffLen := len("Percentage Difference")

	for _, field := range sortedFields {
		fieldLen := len(field)
		if fieldLen > maxFieldLen {
			maxFieldLen = fieldLen
		}

		currentValStr := fmt.Sprintf("%d", int(currentValues[field]))
		//previousValStr := fmt.Sprintf("%d", int(previousValues[field]))
		diffStr := fmt.Sprintf("%d", int(currentValues[field]-previousValues[field]))
		percentageDiffStr := fmt.Sprintf("%.2f%%", computePercentageDiff(currentValues[field], previousValues[field]))

		maxValueLen = max(maxValueLen, len(currentValStr)) //, len(previousValStr))
		maxDiffLen = max(maxDiffLen, len(diffStr))
		maxPercentDiffLen = max(maxPercentDiffLen, len(percentageDiffStr))
	}

	headerFormat := fmt.Sprintf("| %%-%ds | %%-%ds | %%-%ds | %%-%ds | %%-%ds |\n",
		maxFieldLen, maxValueLen, maxValueLen, maxDiffLen, maxPercentDiffLen)
	rowFormat := fmt.Sprintf("| %%-%ds | %%-%dd | %%-%dd | %%-%dd | %%-%ds |\n",
		maxFieldLen, maxValueLen, maxValueLen, maxDiffLen, maxPercentDiffLen)

	fmt.Println(strings.Repeat("-", maxFieldLen+maxValueLen*2+maxDiffLen+maxPercentDiffLen+14))
	fmt.Printf(headerFormat, "Result Type", "Current Build", "Previous Build", "Difference", "Percentage Difference")
	fmt.Println(strings.Repeat("-", maxFieldLen+maxValueLen*2+maxDiffLen+maxPercentDiffLen+14))

	for _, field := range sortedFields {
		currentVal := int(currentValues[field])
		previousVal := int(previousValues[field])
		diff := currentVal - previousVal
		percentageDiff := computePercentageDiff(currentValues[field], previousValues[field])

		fmt.Printf(rowFormat, field, currentVal, previousVal, diff, fmt.Sprintf("%.2f%%", percentageDiff))
	}

	fmt.Println(strings.Repeat("-", maxFieldLen+maxValueLen*2+maxDiffLen+maxPercentDiffLen+14))
}

func computePercentageDiff(currentVal, previousVal float64) float64 {
	if previousVal == 0 {
		return 0.0
	}
	return (currentVal - previousVal) / math.Abs(previousVal) * 100
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

func ToJsonStringFromStringMap(m map[string]interface{}) (string, error) {
	outBytes, err := json.Marshal(m)
	if err == nil {
		return string(outBytes), nil
	}
	return "", err
}

func WriteStrToFile(filePath, data string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
