package plugin

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/sirupsen/logrus"
	"harness-community/drone-test-result-aggregator/plugin/utils"
	"io"
	"os"
	"strings"
	"time"
)

type JacocoAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	utils.DbCredentials
}

type JacocoAggregateData struct {
	utils.ResultBasicInfo
	InstructionTotalSum   float64
	InstructionCoveredSum float64
	InstructionMissedSum  float64
	BranchTotalSum        float64
	BranchCoveredSum      float64
	BranchMissedSum       float64
	LineTotalSum          float64
	LineCoveredSum        float64
	LineMissedSum         float64
	ComplexityTotalSum    float64
	ComplexityCoveredSum  float64
	ComplexityMissedSum   float64
	MethodTotalSum        float64
	MethodCoveredSum      float64
	MethodMissedSum       float64
	ClassTotalSum         float64
	ClassCoveredSum       float64
	ClassMissedSum        float64
}

type Report struct {
	XMLName  xml.Name  `xml:"report"`
	Counters []Counter `xml:"counter"`
	Packages []Package `xml:"package"`
}

type Counter struct {
	Type    string `xml:"type,attr"`
	Missed  int    `xml:"missed,attr"`
	Covered int    `xml:"covered,attr"`
}

type Package struct {
	Name     string    `xml:"name,attr"`
	Counters []Counter `xml:"counter"`
}

type XmlFileReportData struct {
	XMLName struct {
		Space string `json:"Space"`
		Local string `json:"Local"`
	} `json:"XMLName"`
	Counters []struct {
		Type    string `json:"Type"`
		Missed  int    `json:"Missed"`
		Covered int    `json:"Covered"`
	} `json:"Counters"`
	Packages []struct {
		Name     string `json:"Name"`
		Counters []struct {
			Type    string `json:"Type"`
			Missed  int    `json:"Missed"`
			Covered int    `json:"Covered"`
		} `json:"Counters"`
	} `json:"Packages"`
}

func GetNewJacocoAggregator(reportsDir, reportsName, includes,
	dbUrl, dbToken, organization, bucket string) JacocoAggregator {
	return JacocoAggregator{
		ReportsDir:  reportsDir,
		ReportsName: reportsName,
		Includes:    includes,
		DbCredentials: utils.DbCredentials{
			InfluxDBURL:   dbUrl,
			InfluxDBToken: dbToken,
			Organization:  organization,
			Bucket:        bucket,
		},
	}
}

func (j *JacocoAggregator) GetDbUrl() string {
	return j.InfluxDBURL
}

func (j *JacocoAggregator) GetDbToken() string {
	return j.InfluxDBToken
}

func (j *JacocoAggregator) GetDbOrganization() string {
	return j.Organization
}

func (j *JacocoAggregator) GetDbBucket() string {
	return j.Bucket
}

func (j *JacocoAggregator) Aggregate() error {

	reportsRootDir := j.ReportsDir
	patterns := strings.Split(j.Includes, ",")

	xmlFileReportDataList, err := j.GetXmlReportData(reportsRootDir, patterns)
	if err != nil {
		logrus.Println("Error getting xml report data: %v", err)
		return err
	}

	aggregateData := JacocoAggregateData{}
	aggregateData.calculateAggregate(xmlFileReportDataList)

	pipelineId, buildNumber, err := utils.GetPipelineInfo()
	if err != nil {
		logrus.Println("Error getting pipeline info: %v", err)
		return err
	}

	err = j.PersistToInfluxDb(pipelineId, buildNumber, aggregateData)
	if err != nil {
		logrus.Println("Error persisting to influxdb: %v", err)
		return err
	}
	return nil
}

func (j *JacocoAggregator) PersistToInfluxDb(pipelineId, buildNumber string, aggregateData JacocoAggregateData) error {

	aggregateData.Type = utils.JacocoTool
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
			"instruction_total_sum":   aggregateData.InstructionTotalSum,
			"instruction_covered_sum": aggregateData.InstructionCoveredSum,
			"instruction_missed_sum":  aggregateData.InstructionMissedSum,
			"branch_total_sum":        aggregateData.BranchTotalSum,
			"branch_covered_sum":      aggregateData.BranchCoveredSum,
			"branch_missed_sum":       aggregateData.BranchMissedSum,
			"line_total_sum":          aggregateData.LineTotalSum,
			"line_covered_sum":        aggregateData.LineCoveredSum,
			"line_missed_sum":         aggregateData.LineMissedSum,
			"complexity_total_sum":    aggregateData.ComplexityTotalSum,
			"complexity_covered_sum":  aggregateData.ComplexityCoveredSum,
			"complexity_missed_sum":   aggregateData.ComplexityMissedSum,
			"method_total_sum":        aggregateData.MethodTotalSum,
			"method_covered_sum":      aggregateData.MethodCoveredSum,
			"method_missed_sum":       aggregateData.MethodMissedSum,
			"class_total_sum":         aggregateData.ClassTotalSum,
			"class_covered_sum":       aggregateData.ClassCoveredSum,
			"class_missed_sum":        aggregateData.ClassMissedSum,
		},
		time.Now(),
	)

	err := writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		fmt.Println("Error writing point: ", err)
		return err
	}
	fmt.Println("Data persisted successfully to InfluxDB.")
	return nil
}

func (j *JacocoAggregator) GetXmlReportData(
	reportsRootDir string, patterns []string) ([]XmlFileReportData, error) {

	var xmlReportFiles []string
	var xmlFileReportDataList []XmlFileReportData

	for _, pattern := range patterns {
		fmt.Println("pattern ==  ", pattern)
		fmt.Println("reportsRootDir ==  ", reportsRootDir)
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
		fmt.Println("Processing file: ", xmlReportFile)
		report := ParseXMLReport(xmlReportFile)
		reportBytes, err := json.Marshal(report)
		if err != nil {
			logrus.Println("Error marshalling report: %v", err)
		}

		xmlFileReport, err := utils.ToStructFromJsonString[XmlFileReportData](string(reportBytes))
		if err != nil {
			logrus.Println("Error converting json to struct: %v", err)
			return xmlFileReportDataList, err
		}

		xmlFileReportDataList = append(xmlFileReportDataList, xmlFileReport)
	}

	return xmlFileReportDataList, nil
}

func (a *JacocoAggregateData) calculateAggregate(reportsList []XmlFileReportData) {
	for _, report := range reportsList {
		for _, counter := range report.Counters {
			switch counter.Type {
			case "INSTRUCTION":
				addToSum(&a.InstructionTotalSum, &a.InstructionCoveredSum, &a.InstructionMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "BRANCH":
				addToSum(&a.BranchTotalSum, &a.BranchCoveredSum, &a.BranchMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "LINE":
				addToSum(&a.LineTotalSum, &a.LineCoveredSum, &a.LineMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "COMPLEXITY":
				addToSum(&a.ComplexityTotalSum, &a.ComplexityCoveredSum, &a.ComplexityMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "METHOD":
				addToSum(&a.MethodTotalSum, &a.MethodCoveredSum, &a.MethodMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "CLASS":
				addToSum(&a.ClassTotalSum, &a.ClassCoveredSum, &a.ClassMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			}
		}
	}
}

func addToSum(totalSum *float64, coveredSum *float64, missedSum *float64,
	covered float64, missed float64) {
	*totalSum += covered + missed
	*coveredSum += covered
	*missedSum += missed
}

func ParseXMLReport(filename string) Report {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Fatalf("Error opening XML file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		logrus.Fatalf("Error reading XML file: %v", err)
	}

	var report Report
	err = xml.Unmarshal(data, &report)
	if err != nil {
		logrus.Fatalf("Error unmarshalling XML: %v", err)
	}
	return report
}
