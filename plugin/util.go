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
	"os"
	"path/filepath"
	"strings"
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
	getDataMaps func(pipelineId, buildNumber string,
		aggregateData T) (map[string]string, map[string]interface{})) error {

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

func GetPipelineInfo() (string, string, error) {
	pipelineId := os.Getenv(PipeLineIdEnvVar)
	buildNumber := os.Getenv(BuildNumberEnvVar)

	if pipelineId == "" || buildNumber == "" {
		return "", "", fmt.Errorf("PipelineId or BuildNumber not found in the environment")
	}

	return pipelineId, buildNumber, nil
}
