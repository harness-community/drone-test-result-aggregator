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

##  Aggregate Nunit test results
- This step aggregates Nunit test results, stores them in InfluxDB, compares results, and helps understand trends.
- Test results comparison with previous builds can be done using the `compare_build_results` boolean flag.
- When InfluxDB parameters are provided, the plugin will store the test results in InfluxDB. Otherwise, this step is skipped.
- When `compare_build_results` is set to `true`, the plugin will compare the current build results with the previous build results.

### Sample for Aggregate Nunit test results step
```yaml
- step:
    type: Plugin
    name: AggregateNunitTestResultsStep
    identifier: AggregateNunitTestResultsStep
    spec:
      connectorRef: account.harnessImage
      image: plugins/test-results-aggregator:linux-amd64
      settings:
        tool: nunit
        group: suite_01
        reports_dir: /harness/nunit/nunit-multi
        include_pattern: "**/*.trx"
        influxdb_url: http://<influx db url>:8086
        influxdb_token: <+secrets.getValue("influx_db_token")>
        influxdb_org: hns
        influxdb_bucket: hns_test_bucket_02
```

### Sample Nunit result data stored in influxdb

| tableresults | _measurement | _field                 | _value | _start                      | _stop                       | _time                       | buildId | group    | pipelineId          |
|--------------|--------------|------------------------|--------|-----------------------------|-----------------------------|-----------------------------|---------|----------|---------------------|
| 0            | nunit        | aborted               | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 1            | nunit        | completed             | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 2            | nunit        | disconnected          | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 3            | nunit        | error                 | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 4            | nunit        | executed              | 5      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 5            | nunit        | failed                | 1      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 6            | nunit        | inProgress            | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 7            | nunit        | inconclusive          | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 8            | nunit        | notExecuted           | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 9            | nunit        | notRunnable           | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 10           | nunit        | passedButRunAborted   | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 11           | nunit        | pending               | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 12           | nunit        | timeout               | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 13           | nunit        | total                 | 6      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |
| 14           | nunit        | warning               | 0      | 2024-02-05T09:35:27.620Z    | 2025-02-04T15:35:27.620Z    | 2025-02-04T14:55:57.279Z    | 54      | suite_01 | testresultaggregator |

### Test result output as shown in Harness UI
```txt
==================================
  NUnit Test Run Summary
==================================
 Pipeline ID : testresultaggregatorsenthil03linux      
 Build ID : 88                                      
==================================
| Test Category    | Count      |
----------------------------------
| 📁 Total Cases   |         22 |
| ✅ Total Passed  |         21 |
| ❌ Total Failed  |          1 |
| ⏸️ Total Skipped |          0 |
==================================
```

### Test results comparison with previous builds as shown in Harness UI
```txt
Comparison results with previous build:
------------------------------------------------------------------------------------
| Result Type   | Current Build | Previous Build | Difference | Percentage Difference |
------------------------------------------------------------------------------------
| total_cases   | 22            | 22            | 0          | 0.00%                 |
| total_failed  | 1             | 1             | 0          | 0.00%                 |
| total_passed  | 21            | 21            | 0          | 0.00%                 |
| total_skipped | 0             | 0             | 0          | 0.00%                 |
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
