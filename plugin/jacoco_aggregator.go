package plugin

import (
	"encoding/xml"
	"fmt"
)

type JacocoAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	DbCredentials
}

type JacocoAggregateData struct {
	ResultBasicInfo
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
	JacocoAggregateData
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

func GetNewJacocoAggregator(reportsDir, reportsName, includes,
	dbUrl, dbToken, organization, bucket string) JacocoAggregator {
	return JacocoAggregator{
		ReportsDir:  reportsDir,
		ReportsName: reportsName,
		Includes:    includes,
		DbCredentials: DbCredentials{
			InfluxDBURL:   dbUrl,
			InfluxDBToken: dbToken,
			Organization:  organization,
			Bucket:        bucket,
		},
	}
}

func (j *JacocoAggregator) Aggregate() error {
	fmt.Println("Jacoco Aggregator Aggregate")
	err := Aggregate[Report](j.ReportsDir, j.Includes,
		j.DbCredentials.InfluxDBURL, j.DbCredentials.InfluxDBToken,
		j.DbCredentials.Organization, j.DbCredentials.Bucket, JacocoTool,
		CalculateJacocoAggregate, GetJacocoDataMaps)
	return err
}

func CalculateJacocoAggregate(reportsList []Report) Report {

	var xmlFileReportData Report

	for _, report := range reportsList {
		for _, counter := range report.Counters {
			switch counter.Type {
			case "INSTRUCTION":
				addToSum(&xmlFileReportData.InstructionTotalSum, &xmlFileReportData.InstructionCoveredSum, &xmlFileReportData.InstructionMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "BRANCH":
				addToSum(&xmlFileReportData.BranchTotalSum, &xmlFileReportData.BranchCoveredSum, &xmlFileReportData.BranchMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "LINE":
				addToSum(&xmlFileReportData.LineTotalSum, &xmlFileReportData.LineCoveredSum, &xmlFileReportData.LineMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "COMPLEXITY":
				addToSum(&xmlFileReportData.ComplexityTotalSum, &xmlFileReportData.ComplexityCoveredSum, &xmlFileReportData.ComplexityMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "METHOD":
				addToSum(&xmlFileReportData.MethodTotalSum, &xmlFileReportData.MethodCoveredSum, &xmlFileReportData.MethodMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "CLASS":
				addToSum(&xmlFileReportData.ClassTotalSum, &xmlFileReportData.ClassCoveredSum, &xmlFileReportData.ClassMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			}
		}
	}

	return xmlFileReportData
}

func GetJacocoDataMaps(pipelineId, buildNumber string, aggregateData Report) (map[string]string, map[string]interface{}) {
	tagMap := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}
	fieldMap := map[string]interface{}{
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
	}

	return tagMap, fieldMap
}

func addToSum(totalSum *float64, coveredSum *float64, missedSum *float64,
	covered float64, missed float64) {
	*totalSum += covered + missed
	*coveredSum += covered
	*missedSum += missed
}
