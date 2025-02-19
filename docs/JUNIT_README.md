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

##  Aggregate Junit test results
- This step aggregates Junit test results, stores them in InfluxDB, compares results, and helps understand trends.
- Test results comparison with previous builds can be done using the `compare_build_results` boolean flag.
- When InfluxDB parameters are provided, the plugin will store the test results in InfluxDB. Otherwise, this step is skipped.
- When `compare_build_results` is set to `true`, the plugin will compare the current build results with the previous build results.

### Sample for Aggregate Junit test results step
```yaml
- step:
    type: Plugin
    name: AggregateJunitTestResultsStep
    identifier: AggregateJunitTestResultsStep
    spec:
      connectorRef: account.harnessImage
      image: plugins/test-results-aggregator:linux-amd64
      settings:
        tool: junit
        group: suite_01
        reports_dir: /harness/
        include_pattern: "**/TEST*.xml"
        influxdb_url: http://<influx db url>:8086
        influxdb_token: <+secrets.getValue("influx_db_token")>
        influxdb_org: hns
        influxdb_bucket: hns_test_bucket_02
```

### Sample Junit result data stored in influxdb

| results | _measurement | _field   | _value | _start                      | _stop                       | _time                       | build_id | group    | pipeline_id                           |
|---------|-------------|----------|--------|-----------------------------|-----------------------------|-----------------------------|----------|----------|----------------------------------------|
| 0       | junit       | errors   | 9      | 2024-02-05T09:33:28.412Z    | 2025-02-04T15:33:28.412Z    | 2025-02-04T14:56:08.448Z    | 54       | suite_01 | testresultaggregator    |
| 1       | junit       | failures | 12     | 2024-02-05T09:33:28.412Z    | 2025-02-04T15:33:28.412Z    | 2025-02-04T14:56:08.448Z    | 54       | suite_01 | testresultaggregator    |
| 2       | junit       | skipped  | 12     | 2024-02-05T09:33:28.412Z    | 2025-02-04T15:33:28.412Z    | 2025-02-04T14:56:08.448Z    | 54       | suite_01 | testresultaggregator    |
| 3       | junit       | tests    | 11     | 2024-02-05T09:33:28.412Z    | 2025-02-04T15:33:28.412Z    | 2025-02-04T14:56:08.448Z    | 54       | suite_01 | testresultaggregator    |

### Test result output as shown in Harness UI
```txt
=============================================
  JUnit Test Run Summary
=============================================
  Pipeline ID: testresultaggregatorsenthil03linux      
  Build ID: 88                                      
=============================================
| Test Category          | Count             |
---------------------------------------------
| 📁 Total Cases      |       6.00          |
| ✅ Total Passed     |       2.00          |
| ❌ Total Failed     |       2.00          |
| ⏸️ Total Skipped    |       2.00          |
| 🛑 Total Errors     |       0.00          |
=============================================
```

### Test results comparison with previous builds as shown in Harness UI
```txt
Comparison results with previous build:
------------------------------------------------------------------------------------
| Result Type   | Current Build | Previous Build | Difference | Percentage Difference |
------------------------------------------------------------------------------------
| errors_count  | 0             | 0             | 0          | 0.00%                 |
| failed_tests  | 2             | 2             | 0          | 0.00%                 |
| passed_tests  | 2             | 2             | 0          | 0.00%                 |
| skipped_tests | 2             | 2             | 0          | 0.00%                 |
| total_tests   | 6             | 6             | 0          | 0.00%                 |
------------------------------------------------------------------------------------
```

### Exported Environment Variables
| Metric                   | Description |
|--------------------------|-------------|
| **TEST_RESULTS_DATA_FILE** | Contains raw test result metrics, including coverage breakdown and execution data. |
| **TEST_RESULTS_DIFF_FILE** | Stores the differences in test results between builds, helping track regressions and improvements. |


## Community and Support
[Harness Community Slack](https://join.slack.com/t/harnesscommunity/shared_invite/zt-y4hdqh7p-RVuEQyIl5Hcx4Ck8VCvzBw) - Join the #drone slack channel to connect with our engineers and other users running Drone CI.

[Harness Community Forum](https://community.harness.io/) - Ask questions, find answers, and help other users.

[Report and Track A Bug](https://community.harness.io/c/bugs/17) - Find a bug? Please report in our forum under Drone Bugs. Please provide screenshots and steps to reproduce. 

[Events](https://www.meetup.com/harness/) - Keep up to date with Drone events and check out previous events [here](https://www.youtube.com/watch?v=Oq34ImUGcHA&list=PLXsYHFsLmqf3zwelQDAKoVNmLeqcVsD9o).
