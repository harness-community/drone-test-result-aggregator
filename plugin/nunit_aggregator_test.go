package plugin

import (
	"fmt"
	"strings"
	"testing"
)

const NunitTestXml = `<?xml version="1.0" encoding="utf-8"?>
<TestRun id="3e5652b8-f41b-4e07-b3d2-5f28f4f6cc08" name="@ubuntu 2025-01-22 23:28:25" xmlns="http://microsoft.com/schemas/VisualStudio/TeamTest/2010">
  <Times creation="2025-01-22T23:28:25.8095145+05:30" queuing="2025-01-22T23:28:25.8095146+05:30" start="2025-01-22T23:28:25.1581777+05:30" finish="2025-01-22T23:28:25.8173908+05:30" />
  <TestSettings name="default" id="7f157ee1-4e4e-4e47-ae51-e2721f37e8f9">
    <Deployment runDeploymentRoot="_ubuntu_2025-01-22_23_28_25" />
  </TestSettings>
  <Results>
    <UnitTestResult executionId="104b04fb-6049-4b09-af5c-c76330cc6b21" testId="7c13ea3d-4237-3744-118f-4b8a47460eb9" testName="Test2" computerName="ubuntu" duration="00:00:00.0003390" startTime="2025-01-22T23:28:25.6785310+05:30" endTime="2025-01-22T23:28:25.6788696+05:30" testType="13cdc9d9-ddb5-4fa4-a97d-d965ccfc6d4b" outcome="Passed" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" relativeResultsDirectory="104b04fb-6049-4b09-af5c-c76330cc6b21">
      <Output>
        <StdOut>Suite2 - Test2 passed.</StdOut>
      </Output>
    </UnitTestResult>
    <UnitTestResult executionId="e4854018-8d7c-4df8-a7ae-3bb0175b217c" testId="38ee8d2c-88e3-5727-2cbd-76b175ac95c2" testName="Test1" computerName="ubuntu" duration="00:00:00.0166410" startTime="2025-01-22T23:28:25.6618423+05:30" endTime="2025-01-22T23:28:25.6784828+05:30" testType="13cdc9d9-ddb5-4fa4-a97d-d965ccfc6d4b" outcome="Passed" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" relativeResultsDirectory="e4854018-8d7c-4df8-a7ae-3bb0175b217c">
      <Output>
        <StdOut>Suite2 - Test1 passed.</StdOut>
      </Output>
    </UnitTestResult>
    <UnitTestResult executionId="37b76398-bbb7-400b-b492-306c2fb5183f" testId="f9a07328-18a4-0d0d-a9ed-1e4e74e7218b" testName="IgnoredTest" computerName="ubuntu" duration="00:00:00.0015839" startTime="2025-01-22T23:28:25.6586400+05:30" endTime="2025-01-22T23:28:25.6601781+05:30" testType="13cdc9d9-ddb5-4fa4-a97d-d965ccfc6d4b" outcome="NotExecuted" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" relativeResultsDirectory="37b76398-bbb7-400b-b492-306c2fb5183f">
      <Output>
        <StdOut>This test is ignored for demonstration.</StdOut>
        <ErrorInfo>
          <Message>This test is ignored for demonstration.</Message>
        </ErrorInfo>
      </Output>
    </UnitTestResult>
  </Results>
  <TestDefinitions>
    <UnitTest name="IgnoredTest" storage="/opt/hns/test-resources/test-result-aggregator/nunit/nunit-multi/src/tests/bin/debug/net8.0/tests.dll" id="f9a07328-18a4-0d0d-a9ed-1e4e74e7218b">
      <Execution id="37b76398-bbb7-400b-b492-306c2fb5183f" />
      <TestMethod codeBase="/opt/hns/test-resources/test-result-aggregator/nunit/nunit-multi/src/Tests/bin/Debug/net8.0/Tests.dll" adapterTypeName="executor://nunit3testexecutor/" className="nunit_multi.Tests.Suite2.Suite2Tests" name="IgnoredTest" />
    </UnitTest>
    <UnitTest name="Test1" storage="/opt/hns/test-resources/test-result-aggregator/nunit/nunit-multi/src/tests/bin/debug/net8.0/tests.dll" id="38ee8d2c-88e3-5727-2cbd-76b175ac95c2">
      <Execution id="e4854018-8d7c-4df8-a7ae-3bb0175b217c" />
      <TestMethod codeBase="/opt/hns/test-resources/test-result-aggregator/nunit/nunit-multi/src/Tests/bin/Debug/net8.0/Tests.dll" adapterTypeName="executor://nunit3testexecutor/" className="nunit_multi.Tests.Suite2.Suite2Tests" name="Test1" />
    </UnitTest>
    <UnitTest name="Test2" storage="/opt/hns/test-resources/test-result-aggregator/nunit/nunit-multi/src/tests/bin/debug/net8.0/tests.dll" id="7c13ea3d-4237-3744-118f-4b8a47460eb9">
      <Execution id="104b04fb-6049-4b09-af5c-c76330cc6b21" />
      <TestMethod codeBase="/opt/hns/test-resources/test-result-aggregator/nunit/nunit-multi/src/Tests/bin/Debug/net8.0/Tests.dll" adapterTypeName="executor://nunit3testexecutor/" className="nunit_multi.Tests.Suite2.Suite2Tests" name="Test2" />
    </UnitTest>
  </TestDefinitions>
  <TestEntries>
    <TestEntry testId="7c13ea3d-4237-3744-118f-4b8a47460eb9" executionId="104b04fb-6049-4b09-af5c-c76330cc6b21" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" />
    <TestEntry testId="38ee8d2c-88e3-5727-2cbd-76b175ac95c2" executionId="e4854018-8d7c-4df8-a7ae-3bb0175b217c" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" />
    <TestEntry testId="f9a07328-18a4-0d0d-a9ed-1e4e74e7218b" executionId="37b76398-bbb7-400b-b492-306c2fb5183f" testListId="8c84fa94-04c1-424b-9868-57a2d4851a1d" />
  </TestEntries>
  <TestLists>
    <TestList name="Results Not in a List" id="8c84fa94-04c1-424b-9868-57a2d4851a1d" />
    <TestList name="All Loaded Results" id="19431567-8539-422a-85d7-44ee4e166bda" />
  </TestLists>
  <ResultSummary outcome="Completed">
    <Counters total="3" executed="2" passed="2" failed="0" error="0" timeout="0" aborted="0" inconclusive="0" passedButRunAborted="0" notRunnable="0" notExecuted="0" disconnected="0" warning="0" completed="0" inProgress="0" pending="0" />
    <Output>
      <StdOut>NUnit Adapter 4.6.0.0: Test execution started
Running selected tests in /opt/hns/test-resources/test-result-aggregator/nunit/nunit-multi/src/Tests/bin/Debug/net8.0/Tests.dll
   NUnit3TestExecutor discovered 3 of 3 NUnit test cases using Current Discovery mode, Non-Explicit run
IgnoredTest: This test is ignored for demonstration.
Test1: Suite2 - Test1 passed.
Test2: Suite2 - Test2 passed.
NUnit Adapter 4.6.0.0: Test execution complete
Test 'IgnoredTest' was skipped in the test run.
</StdOut>
    </Output>
  </ResultSummary>
</TestRun>`

func TestNunitAggregator(t *testing.T) {
	tagsMap, fieldsMap := MockAggregate[TestRun](NunitTestXml,
		CalculateNugetAggregate, GetNunitDataMaps)

	expectedTagsMap := map[string]string{
		"buildId":    mockBuildNumber,
		"pipelineId": mockPipelineId,
	}

	expectedFieldsMap := map[string]interface{}{
		"outcome":  "Completed",
		"total":    3,
		"executed": 2,
		"passed":   2,
		"failed":   0,
	}

	for k := range expectedTagsMap {
		if tagsMap[k] != expectedTagsMap[k] {
			t.Errorf("===> Mismatch in TagsMap for key %q: got %v, expected %v",
				k, tagsMap[k], expectedTagsMap[k])
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

func TestComputeNunitBuildResultDifferences(t *testing.T) {
	currentValues := map[string]float64{
		"total":               100,
		"executed":            95,
		"passed":              90,
		"failed":              5,
		"error":               1,
		"timeout":             0,
		"aborted":             0,
		"inconclusive":        2,
		"passedButRunAborted": 0,
		"notRunnable":         1,
		"notExecuted":         1,
		"disconnected":        0,
		"warning":             3,
		"completed":           95,
		"inProgress":          2,
		"pending":             3,
	}

	previousValues := map[string]float64{
		"total":               100,
		"executed":            95,
		"passed":              90,
		"failed":              5,
		"error":               1,
		"timeout":             0,
		"aborted":             0,
		"inconclusive":        2,
		"passedButRunAborted": 0,
		"notRunnable":         1,
		"notExecuted":         1,
		"disconnected":        0,
		"warning":             3,
		"completed":           95,
		"inProgress":          2,
		"pending":             3,
	}

	expectedResultDiffs := []ResultDiff{}
	for field, currValue := range currentValues {
		prevValue := previousValues[field]

		expectedResultDiffs = append(expectedResultDiffs, ResultDiff{
			FieldName:            field,
			CurrentBuildValue:    currValue,
			PreviousBuildValue:   prevValue,
			Difference:           currValue - prevValue,
			PercentageDifference: 0,
			IsCompareValid:       true,
		})
	}

	result := ComputeBuildResultDifferences(currentValues, previousValues)

	expectedCsvRows := []string{
		"ResultType, CurrentBuild, PreviousBuild, Difference, PercentageDifference",
		"total, 100.00, 100.00, 0.00, 0.00",
		"disconnected, 0.00, 0.00, 0.00, 0.00",
		"notExecuted, 1.00, 1.00, 0.00, 0.00",
		"warning, 3.00, 3.00, 0.00, 0.00",
		"aborted, 0.00, 0.00, 0.00, 0.00",
		"timeout, 0.00, 0.00, 0.00, 0.00",
		"notRunnable, 1.00, 1.00, 0.00, 0.00",
		"pending, 3.00, 3.00, 0.00, 0.00",
		"passedButRunAborted, 0.00, 0.00, 0.00, 0.00",
		"inProgress, 2.00, 2.00, 0.00, 0.00",
		"completed, 95.00, 95.00, 0.00, 0.00",
		"error, 1.00, 1.00, 0.00, 0.00",
		"failed, 5.00, 5.00, 0.00, 0.00",
		"executed, 95.00, 95.00, 0.00, 0.00",
		"inconclusive, 2.00, 2.00, 0.00, 0.00",
		"passed, 90.00, 90.00, 0.00, 0.00",
	}

	for _, expectedRow := range expectedCsvRows {
		found := strings.Contains(result, expectedRow)
		if !found {
			t.Errorf("Expected row not found in result: %q", expectedRow)
		}
	}
}
