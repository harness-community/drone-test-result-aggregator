package plugin

import (
	"encoding/xml"
	"fmt"
	"github.com/sirupsen/logrus"
)

const mockPipelineId = "pipelineId001"
const mockBuildNumber = "201"

func MockAggregate[T any](xmlReportStr string,
	calculateAggregate func(testNgAggregatorList []T) T,
	getDataMaps func(pipelineId, buildNumber string,
		aggregateData T) (map[string]string, map[string]interface{})) (
	map[string]string, map[string]interface{}) {

	aggregatorList := MockParseXmlReport[T](xmlReportStr)

	totalAggregate := calculateAggregate(aggregatorList)
	fmt.Println("Total Aggregate: ", totalAggregate)

	tagsMap, fieldsMap := getDataMaps(mockPipelineId, mockBuildNumber, totalAggregate)

	return tagsMap, fieldsMap
}

func MockParseXmlReport[T any](xmlStr string) []T {
	data := []byte(xmlStr)
	var report T
	err := xml.Unmarshal(data, &report)
	if err != nil {
		fmt.Printf("Error unmarshalling XML: %v", err)
		logrus.Fatalf("Error unmarshalling XML: %v", err)
	}
	return []T{report}
}
