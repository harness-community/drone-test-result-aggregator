A plugin to aggregate test results, store in influx DB, compare results and understand trends.

Run the following script to install git-leaks support to this repo.
```
chmod +x ./git-hooks/install.sh
./git-hooks/install.sh
```

## Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t plugins/test-result-aggregator  -f docker/Dockerfile .
```

##  Aggregate Jacoco test results
- This step aggregates Jacoco test results, stores them in InfluxDB, compares results, and helps understand trends.
- Test results comparison with previous builds can be done using the `compare_build_results` boolean flag.
- When InfluxDB parameters are provided, the plugin will store the test results in InfluxDB. Otherwise, this step is skipped.
- When `compare_build_results` is set to `true`, the plugin will compare the current build results with the previous build results.


### Sample for Aggregate Jacoco test results step
```yaml
- step:
    type: Plugin
    name: AggregateJacocoTestResultsStep
    identifier: AggregateJacocoTestResultsStep
    spec:
      connectorRef: account.harnessImage
      image: plugins/test-results-aggregator:linux-amd64
      settings:
        tool: jacoco
        group: suite_01
        reports_dir: /harness/
        include_pattern: "**/jacoco*.xml"
        influxdb_url: http://<influx db url>:8086
        influxdb_token: <+secrets.getValue("influx_db_token")>
        influxdb_org: hns
        influxdb_bucket: hns_test_bucket_02
```

### Sample Jacoco result data stored in influxdb

| tableresults | _measurement | _field                 | _value | _start                      | _stop                       | _time                       | buildId | group    | pipelineId                            |
|-------------|--------------|------------------------|--------|-----------------------------|-----------------------------|-----------------------------|---------|----------|----------------------------------------|
| 0           | jacoco       | branch_covered_sum    | 114    | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 1           | jacoco       | branch_missed_sum     | 2      | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 2           | jacoco       | branch_total_sum      | 116    | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 3           | jacoco       | class_covered_sum     | 10     | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 4           | jacoco       | class_missed_sum      | 0      | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 5           | jacoco       | class_total_sum       | 10     | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 6           | jacoco       | complexity_covered_sum | 140    | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 7           | jacoco       | complexity_missed_sum | 2      | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 8           | jacoco       | complexity_total_sum  | 142    | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 9           | jacoco       | instruction_covered_sum | 1240 | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 10          | jacoco       | instruction_missed_sum | 0      | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 11          | jacoco       | instruction_total_sum | 1240   | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 12          | jacoco       | line_covered_sum      | 244    | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 13          | jacoco       | line_missed_sum       | 0      | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 14          | jacoco       | line_total_sum        | 244    | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 15          | jacoco       | method_covered_sum    | 84     | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 16          | jacoco       | method_missed_sum     | 0      | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |
| 17          | jacoco       | method_total_sum      | 84     | 2024-02-05T09:03:22.181Z    | 2025-02-04T15:03:22.181Z    | 2025-02-04T14:55:47.049Z    | 54      | suite_01 | testresultaggregator    |

### Test result output as shown in Harness UI
```txt
===================================================================
  Jacoco Code Coverage Summary                                                              
===================================================================
  Pipeline ID         : testresultaggregatorsenthil03linux                                
  Build ID            : 88                                                                
===================================================================
| Coverage Type             | Total      | Covered    | Missed     |
-------------------------------------------------------------------
| ✅ Instruction Coverage    |    1240.00 |    1240.00 |       0.00 |
| ✅ Branch Coverage         |     116.00 |     114.00 |       2.00 |
| ✅ Line Coverage           |     244.00 |     244.00 |       0.00 |
| ✅ Complexity Coverage     |     142.00 |     140.00 |       2.00 |
| ✅ Method Coverage         |      84.00 |      84.00 |       0.00 |
| ✅ Class Coverage          |      10.00 |      10.00 |       0.00 |
===================================================================
```

### Test results comparison with previous builds as shown in Harness UI
```txt
Comparison results with previous build:
----------------------------------------------------------------------------------------------
| Result Type             | Current Build | Previous Build | Difference | Percentage Difference |
----------------------------------------------------------------------------------------------
| branch_covered_sum      | 114           | 114           | 0          | 0.00%                 |
| branch_missed_sum       | 2             | 2             | 0          | 0.00%                 |
| branch_total_sum        | 116           | 116           | 0          | 0.00%                 |
| class_covered_sum       | 10            | 10            | 0          | 0.00%                 |
| class_missed_sum        | 0             | 0             | 0          | 0.00%                 |
| class_total_sum         | 10            | 10            | 0          | 0.00%                 |
| complexity_covered_sum  | 140           | 140           | 0          | 0.00%                 |
| complexity_missed_sum   | 2             | 2             | 0          | 0.00%                 |
| complexity_total_sum    | 142           | 142           | 0          | 0.00%                 |
| instruction_covered_sum | 1240          | 1240          | 0          | 0.00%                 |
| instruction_missed_sum  | 0             | 0             | 0          | 0.00%                 |
| instruction_total_sum   | 1240          | 1240          | 0          | 0.00%                 |
| line_covered_sum        | 244           | 244           | 0          | 0.00%                 |
| line_missed_sum         | 0             | 0             | 0          | 0.00%                 |
| line_total_sum          | 244           | 244           | 0          | 0.00%                 |
| method_covered_sum      | 84            | 84            | 0          | 0.00%                 |
| method_missed_sum       | 0             | 0             | 0          | 0.00%                 |
| method_total_sum        | 84            | 84            | 0          | 0.00%                 |
----------------------------------------------------------------------------------------------
```
### Exported Environment Variables
| Metric                   | Description                                                                                                                                 |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------|
| **BRANCH_COVERAGE**      | Measures the percentage of executed branches in conditional statements. A high value indicates good decision coverage.                      |
| **COMPLEXITY_COVERAGE**  | Represents the percentage of executed paths within methods, considering code complexity. Higher coverage means more tested execution paths. |
| **METHOD_COVERAGE**      | Indicates how many methods were executed during tests. A value of 100% means all methods were tested at least once.                         |
| **LINE_COVERAGE**        | Shows the percentage of executed lines of code. 100% means all lines were covered by tests.                                                 |
| **CLASS_COVERAGE**       | Reflects how many classes have been fully executed by the tests. 100% indicates all classes were tested.                                    |
| **INSTRUCTION_COVERAGE** | Measures the number of executed bytecode instructions, ensuring thorough test execution.                                                    |
| **TEST_RESULTS_DATA_FILE** | File storage path, contains raw test result metrics, including coverage breakdown and execution data.                                       |
| **TEST_RESULTS_DIFF_FILE** | File storage path, Stores the differences in test results between builds, helping track regressions and improvements.                                          |


## Community and Support
[Harness Community Slack](https://join.slack.com/t/harnesscommunity/shared_invite/zt-y4hdqh7p-RVuEQyIl5Hcx4Ck8VCvzBw) - Join the #drone slack channel to connect with our engineers and other users running Drone CI.

[Harness Community Forum](https://community.harness.io/) - Ask questions, find answers, and help other users.

[Report and Track A Bug](https://community.harness.io/c/bugs/17) - Find a bug? Please report in our forum under Drone Bugs. Please provide screenshots and steps to reproduce. 

[Events](https://www.meetup.com/harness/) - Keep up to date with Drone events and check out previous events [here](https://www.youtube.com/watch?v=Oq34ImUGcHA&list=PLXsYHFsLmqf3zwelQDAKoVNmLeqcVsD9o).
