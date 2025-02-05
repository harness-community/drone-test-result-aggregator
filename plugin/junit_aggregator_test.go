package plugin

import (
	"fmt"
	"reflect"
	"testing"
)

func TestJunitAggregator_Aggregate(t *testing.T) {

	tagsMap, fieldsMap := MockAggregate[JunitAggregatorData](JunitReportXml,
		CalculateJunitAggregate, GetJunitDataMaps)

	expectedTagsMap := map[string]string{
		"build_id":    mockBuildNumber,
		"pipeline_id": mockPipelineId,
	}

	expectedFieldsMap := map[string]interface{}{
		"errors":   7,
		"failures": 9,
		"skipped":  8,
		"tests":    6,
	}

	for k := range expectedTagsMap {
		if tagsMap[k] != expectedTagsMap[k] {
			t.Errorf("Mismatch in TagsMap for key %q: got %v, expected %v", k, tagsMap[k], expectedTagsMap[k])
		}
	}

	for k := range expectedFieldsMap {
		gotVal := fmt.Sprintf("%v", fieldsMap[k])
		expectedVal := fmt.Sprintf("%v", expectedFieldsMap[k])
		if gotVal != expectedVal {
			t.Errorf("Mismatch in FieldsMap for key %q: got %v, expected %v", k, gotVal, expectedVal)
		}
	}

}

const JunitReportXml = `<?xml version="1.0" encoding="UTF-8"?>
<testsuite xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="https://maven.apache.org/surefire/maven-surefire-plugin/xsd/surefire-test-report.xsd" version="3.0.2" name="com.example.project.CalculatorTests" time="0.168" tests="6" errors="7" skipped="8" failures="9">
 <properties>
   <property name="os.name" value="Linux"/>
   <property name="user.name" value="hns"/>
   <property name="path.separator" value=":"/>
   <property name="sun.io.unicode.encoding" value="UnicodeLittle"/>
   <property name="java.class.version" value="52.0"/>
 </properties>
 <testcase name="addsTwoNumbers" classname="com.example.project.CalculatorTests" time="0.034"/>
 <testcase name="add(int, int, int)[1]" classname="com.example.project.CalculatorTests" time="0.022"/>
 <testcase name="add(int, int, int)[2]" classname="com.example.project.CalculatorTests" time="0.001"/>
 <testcase name="add(int, int, int)[3]" classname="com.example.project.CalculatorTests" time="0.002"/>
 <testcase name="add(int, int, int)[4]" classname="com.example.project.CalculatorTests" time="0.001"/>
</testsuite>`

func TestGetJunitDataMaps(t *testing.T) {
	pipelineId := "test_pipeline"
	buildNumber := "5"

	aggregateData := JunitAggregatorData{
		Name:     "JUnit Test Suite",
		Tests:    100,
		Skipped:  5,
		Failures: 2,
		Errors:   1,
	}

	expectedTagsMap := map[string]string{
		"pipeline_id": pipelineId,
		"build_id":    buildNumber,
		"name":        "JUnit Test Suite",
		"type":        "",
		"status":      "",
	}

	expectedFieldsMap := map[string]interface{}{
		"tests":    100,
		"skipped":  5,
		"failures": 2,
		"errors":   1,
	}

	tagsMap, fieldsMap := GetJunitDataMaps(pipelineId, buildNumber, aggregateData)
	if !reflect.DeepEqual(tagsMap, expectedTagsMap) {
		t.Errorf("tagsMap mismatch.\nExpected: %+v\nGot: %+v", expectedTagsMap, tagsMap)
	}

	if !reflect.DeepEqual(fieldsMap, expectedFieldsMap) {
		t.Errorf("fieldsMap mismatch.\nExpected: %+v\nGot: %+v", expectedFieldsMap, fieldsMap)
	}
}
