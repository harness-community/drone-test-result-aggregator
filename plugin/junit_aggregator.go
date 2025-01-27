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

type JunitAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	DbCredentials
}

type JunitAggregatorData struct {
	ResultBasicInfo
	XMLName   xml.Name `xml:"testsuite"`
	Name      string   `xml:"name,attr"`
	Tests     int      `xml:"tests,attr"`
	Skipped   int      `xml:"skipped,attr"`
	Failures  int      `xml:"failures,attr"`
	Errors    int      `xml:"errors,attr"`
	Timestamp string   `xml:"timestamp,attr,omitempty"`
	Hostname  string   `xml:"hostname,attr,omitempty"`
	Time      float64  `xml:"time,attr"`
	Version   string   `xml:"version,attr,omitempty"`
	Schema    string   `xml:"xsi:noNamespaceSchemaLocation,attr,omitempty"`
}

func GetNewJunitAggregator(
	reportsDir, reportsName, includes, dbUrl, dbToken, dbOrg, dbBucket string) *JunitAggregator {
	return &JunitAggregator{
		ReportsDir:  reportsDir,
		ReportsName: reportsName,
		Includes:    includes,
		DbCredentials: DbCredentials{
			InfluxDBURL:   dbUrl,
			InfluxDBToken: dbToken,
			Organization:  dbOrg,
			Bucket:        dbBucket,
		},
	}
}

func (j *JunitAggregator) GetDbUrl() string {
	return j.DbCredentials.InfluxDBURL
}

func (j *JunitAggregator) GetDbToken() string {
	return j.DbCredentials.InfluxDBToken
}

func (j *JunitAggregator) GetDbOrganization() string {
	return j.DbCredentials.Organization
}

func (j *JunitAggregator) GetDbBucket() string {
	return j.DbCredentials.Bucket
}
func (j *JunitAggregator) Aggregate() error {

	reportsRootDir := j.ReportsDir
	patterns := strings.Split(j.Includes, ",")

	junitAggregatorDataList, err := GetXmlReportData[JunitAggregatorData](reportsRootDir, patterns)
	if err != nil {
		fmt.Println("Error getting xml report data: ", err.Error())
		return err
	}

	totalAggregate := j.calculateAggregate(junitAggregatorDataList)
	fmt.Println("Total Aggregate: ", totalAggregate)

	pipelineId, buildNumber, err := GetPipelineInfo()
	if err != nil {
		fmt.Println("Error getting pipeline info: ", err.Error())
		return err
	}

	err = j.PersistToInfluxDb(pipelineId, buildNumber, totalAggregate)
	return nil
}

func (j *JunitAggregator) PersistToInfluxDb(pipelineId, buildNumber string, aggregateData JunitAggregatorData) error {
	aggregateData.Type = JunitTool
	aggregateData.PipelineId = pipelineId
	aggregateData.BuildId = buildNumber

	client := influxdb2.NewClient(j.GetDbUrl(), j.GetDbToken())
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(j.GetDbOrganization(), j.GetDbBucket())
	point := influxdb2.NewPoint(
		aggregateData.Type,
		map[string]string{
			"pipeline_id": pipelineId,
			"build_id":    buildNumber,
			"name":        aggregateData.Name,
			"type":        aggregateData.Type,
			"status":      aggregateData.Status,
		},
		map[string]interface{}{
			"tests":    aggregateData.Tests,
			"skipped":  aggregateData.Skipped,
			"failures": aggregateData.Failures,
			"errors":   aggregateData.Errors,
		},
		time.Now())
	err := writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		fmt.Println("Error writing point: ", err)
		return err
	}
	fmt.Println("Data persisted successfully to InfluxDB.")
	return nil
}

func (_ *JunitAggregator) calculateAggregate(reportsList []JunitAggregatorData) JunitAggregatorData {
	var tmpJunitAggregatorData JunitAggregatorData

	for _, report := range reportsList {
		tmpJunitAggregatorData.Tests += report.Tests
		tmpJunitAggregatorData.Skipped += report.Skipped
		tmpJunitAggregatorData.Failures += report.Failures
		tmpJunitAggregatorData.Errors += report.Errors
	}

	return tmpJunitAggregatorData
}

func GetXmlReportData[T any](reportsRootDir string, patterns []string) ([]T, error) {

	fmt.Println("GetXmlReportData: reportsRootDir ==  ", reportsRootDir)

	var xmlReportFiles []string
	var xmlFileReportDataList []T

	for _, pattern := range patterns {
		fmt.Println("junit: pattern ==  ", pattern)
		fmt.Println("junit: reportsRootDir ==  ", reportsRootDir)
		tmpReportDir := os.DirFS(reportsRootDir)
		relPattern := strings.TrimPrefix(pattern, reportsRootDir+"/")
		filesList, err := doublestar.Glob(tmpReportDir, relPattern)
		if err != nil {
			logrus.Println("Include patterns not found ", err.Error())
			return xmlFileReportDataList, err
		}
		xmlReportFiles = append(xmlReportFiles, filesList...)
	}

	fmt.Println("xmlReportFiles ==  ", xmlReportFiles)
	fmt.Println("len(xmlReportFiles) ", len(xmlReportFiles))

	for _, xmlReportFile := range xmlReportFiles {
		fmt.Println("Processing junit result file: ", xmlReportFile)
		tmpXmlReportFile := filepath.Join(reportsRootDir, xmlReportFile)
		report := ParseXmlReport[T](tmpXmlReportFile)
		reportBytes, err := json.Marshal(report)
		if err != nil {
			fmt.Println("Error marshalling report: %v", err)
			logrus.Println("Error marshalling report: %v", err)
		}

		xmlFileReport, err := ToStructFromJsonString[T](string(reportBytes))
		if err != nil {
			fmt.Println("Error converting json to struct: %v", err)
			logrus.Println("Error converting json to struct: %v", err)
			return xmlFileReportDataList, err
		}

		xmlFileReportDataList = append(xmlFileReportDataList, xmlFileReport)
	}

	return xmlFileReportDataList, nil
}

func ParseXmlReport[T any](filename string) T {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening XML file: ", err)
		logrus.Fatalf("Error opening XML file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading XML file ", err)
	}

	var report T
	err = xml.Unmarshal(data, &report)
	if err != nil {
		fmt.Println("Error unmarshalling XML: %v", err)
	}
	return report
}
