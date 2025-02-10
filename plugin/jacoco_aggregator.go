package plugin

import (
	"encoding/xml"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
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

func (j *JacocoAggregator) Aggregate(groupName string) error {

	logrus.Println("Jacoco Aggregator Aggregate")
	tagsMap, fieldsMap, err := Aggregate[Report](j.ReportsDir, j.Includes,
		j.DbCredentials.InfluxDBURL, j.DbCredentials.InfluxDBToken,
		j.DbCredentials.Organization, j.DbCredentials.Bucket, JacocoTool, groupName,
		CalculateJacocoAggregate, GetJacocoDataMaps, ShowJacocoStats)

	err = ExportJacocoOutputVars(tagsMap, fieldsMap)
	if err != nil {
		logrus.Errorf("Error exporting Jacoco coverage metrics: %v", err)
		return err
	}

	return nil
}

func ExportJacocoOutputVars(tagsMap map[string]string, fieldsMap map[string]interface{}) error {

	instructionCoveragePercentage := CalculatePercentage(int(fieldsMap["instruction_covered_sum"].(float64)),
		int(fieldsMap["instruction_missed_sum"].(float64)))
	branchCoveragePercentage := CalculatePercentage(int(fieldsMap["branch_covered_sum"].(float64)),
		int(fieldsMap["branch_missed_sum"].(float64)))
	lineCoveragePercentage := CalculatePercentage(int(fieldsMap["line_covered_sum"].(float64)),
		int(fieldsMap["line_missed_sum"].(float64)))
	complexityCoverage := CalculatePercentage(int(fieldsMap["complexity_covered_sum"].(float64)),
		int(fieldsMap["complexity_missed_sum"].(float64)))
	methodCoveragePercentage := CalculatePercentage(int(fieldsMap["method_covered_sum"].(float64)),
		int(fieldsMap["method_missed_sum"].(float64)))
	classCoveragePercentage := CalculatePercentage(int(fieldsMap["class_covered_sum"].(float64)),
		int(fieldsMap["class_missed_sum"].(float64)))

	outputVarsMap := map[string]interface{}{
		"INSTRUCTION_COVERAGE": instructionCoveragePercentage,
		"BRANCH_COVERAGE":      branchCoveragePercentage,
		"LINE_COVERAGE":        lineCoveragePercentage,
		"COMPLEXITY_COVERAGE":  complexityCoverage,
		"METHOD_COVERAGE":      methodCoveragePercentage,
		"CLASS_COVERAGE":       classCoveragePercentage,
	}

	for key, value := range outputVarsMap {
		err := WriteToEnvVariable(key, value)
		if err != nil {
			logrus.Errorf("Error writing to env variable: %v", err)
			return err
		}
	}
	return nil
}

func CalculateJacocoAggregate(reportsList []Report) Report {

	var xmlFileReportData Report

	for _, report := range reportsList {
		for _, counter := range report.Counters {
			switch counter.Type {
			case "INSTRUCTION":
				addToSum(&xmlFileReportData.InstructionTotalSum, &xmlFileReportData.InstructionCoveredSum,
					&xmlFileReportData.InstructionMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "BRANCH":
				addToSum(&xmlFileReportData.BranchTotalSum,
					&xmlFileReportData.BranchCoveredSum, &xmlFileReportData.BranchMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "LINE":
				addToSum(&xmlFileReportData.LineTotalSum,
					&xmlFileReportData.LineCoveredSum, &xmlFileReportData.LineMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "COMPLEXITY":
				addToSum(&xmlFileReportData.ComplexityTotalSum,
					&xmlFileReportData.ComplexityCoveredSum, &xmlFileReportData.ComplexityMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "METHOD":
				addToSum(&xmlFileReportData.MethodTotalSum,
					&xmlFileReportData.MethodCoveredSum, &xmlFileReportData.MethodMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			case "CLASS":
				addToSum(&xmlFileReportData.ClassTotalSum,
					&xmlFileReportData.ClassCoveredSum, &xmlFileReportData.ClassMissedSum,
					float64(counter.Covered), float64(counter.Missed))
			}
		}
	}

	return xmlFileReportData
}

func GetJacocoDataMaps(pipelineId, buildNumber string,
	aggregateData Report) (map[string]string, map[string]interface{}) {
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

func CalculatePercentage(covered, missed int) float64 {
	total := covered + missed
	if total == 0 {
		return 0.0
	}
	return (float64(covered) / float64(total)) * 100
}

func ShowJacocoStats(tags map[string]string, fields map[string]interface{}) error {
	border := "==================================================================="
	separator := "-------------------------------------------------------------------"

	table := []string{
		border,
		fmt.Sprintf("  %-48s %-40s ", "Jacoco Code Coverage Summary ", ""),
		border,
		fmt.Sprintf("  %-20s: %-65s ", "Pipeline ID", tags["pipelineId"]),
		fmt.Sprintf("  %-20s: %-65s ", "Build ID", tags["buildId"]),
		border,
		fmt.Sprintf("| %-25s | %-10s | %-10s | %-10s |", "Coverage Type", "Total", "Covered", "Missed"),
		separator,
		fmt.Sprintf("| %-25s | %10.2f | %10.2f | %10.2f |", "✅ Instruction Coverage",
			fields["instruction_total_sum"], fields["instruction_covered_sum"], fields["instruction_missed_sum"]),
		fmt.Sprintf("| %-25s | %10.2f | %10.2f | %10.2f |", "✅ Branch Coverage",
			fields["branch_total_sum"], fields["branch_covered_sum"], fields["branch_missed_sum"]),
		fmt.Sprintf("| %-25s | %10.2f | %10.2f | %10.2f |", "✅ Line Coverage",
			fields["line_total_sum"], fields["line_covered_sum"], fields["line_missed_sum"]),
		fmt.Sprintf("| %-25s | %10.2f | %10.2f | %10.2f |", "✅ Complexity Coverage",
			fields["complexity_total_sum"], fields["complexity_covered_sum"], fields["complexity_missed_sum"]),
		fmt.Sprintf("| %-25s | %10.2f | %10.2f | %10.2f |", "✅ Method Coverage",
			fields["method_total_sum"], fields["method_covered_sum"], fields["method_missed_sum"]),
		fmt.Sprintf("| %-25s | %10.2f | %10.2f | %10.2f |", "✅ Class Coverage",
			fields["class_total_sum"], fields["class_covered_sum"], fields["class_missed_sum"]),
		border,
	}

	fmt.Println(strings.Join(table, "\n"))
	return nil
}
