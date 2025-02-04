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
This step aggregates Jacoco test results, stores in influx DB, compares results and helps understand trends. <br>
Test results comparison with previous builds can be done using compare_build_results boolean flag. <br>

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


## Community and Support
[Harness Community Slack](https://join.slack.com/t/harnesscommunity/shared_invite/zt-y4hdqh7p-RVuEQyIl5Hcx4Ck8VCvzBw) - Join the #drone slack channel to connect with our engineers and other users running Drone CI.

[Harness Community Forum](https://community.harness.io/) - Ask questions, find answers, and help other users.

[Report and Track A Bug](https://community.harness.io/c/bugs/17) - Find a bug? Please report in our forum under Drone Bugs. Please provide screenshots and steps to reproduce. 

[Events](https://www.meetup.com/harness/) - Keep up to date with Drone events and check out previous events [here](https://www.youtube.com/watch?v=Oq34ImUGcHA&list=PLXsYHFsLmqf3zwelQDAKoVNmLeqcVsD9o).
