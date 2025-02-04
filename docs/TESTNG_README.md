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

##  Aggregate Testng test results
This step aggregates Testng test results, stores in influx DB, compares results and helps understand trends. <br>
Test results comparison with previous builds can be done using compare_build_results boolean flag.<br>

### Aggregate Testng test results, store in influx DB, compare results and understand trends
```yaml
- step:
    type: Plugin
    name: AggregateTestngTestResultsStep
    identifier: AggregateTestngTestResultsStep
    spec:
      connectorRef: account.harnessImage
      image: plugins/test-results-aggregator:linux-amd64
      settings:
        tool: testng
        group: suite_01
        reports_dir: /harness/test-ng
        include_pattern: "**/testng-results*.xml"
        influxdb_url: http://<influx db url>:8086
        influxdb_token: <+secrets.getValue("influx_db_token")>
        influxdb_org: hns
        influxdb_bucket: hns_test_bucket_02
```

### Sample Testng result data stored in influxdb
| results | _measurement | _field  | _value | _start                      | _stop                       | _time                       | buildId | group    | pipelineId                           |
|---------|-------------|---------|--------|-----------------------------|-----------------------------|-----------------------------|---------|----------|---------------------------------------|
| 0       | testng      | failed  | 2      | 2024-02-05T09:47:49.064Z    | 2025-02-04T15:47:49.064Z    | 2025-02-04T14:56:19.591Z    | 54      | suite_01 | testresultaggregator   |
| 1       | testng      | ignored | 0      | 2024-02-05T09:47:49.064Z    | 2025-02-04T15:47:49.064Z    | 2025-02-04T14:56:19.591Z    | 54      | suite_01 | testresultaggregator   |
| 2       | testng      | passed  | 6      | 2024-02-05T09:47:49.064Z    | 2025-02-04T15:47:49.064Z    | 2025-02-04T14:56:19.591Z    | 54      | suite_01 | testresultaggregator   |
| 3       | testng      | skipped | 2      | 2024-02-05T09:47:49.064Z    | 2025-02-04T15:47:49.064Z    | 2025-02-04T14:56:19.591Z    | 54      | suite_01 | testresultaggregator   |
| 4       | testng      | total   | 10     | 2024-02-05T09:47:49.064Z    | 2025-02-04T15:47:49.064Z    | 2025-02-04T14:56:19.591Z    | 54      | suite_01 | testresultaggregator   |


## Community and Support
[Harness Community Slack](https://join.slack.com/t/harnesscommunity/shared_invite/zt-y4hdqh7p-RVuEQyIl5Hcx4Ck8VCvzBw) - Join the #drone slack channel to connect with our engineers and other users running Drone CI.

[Harness Community Forum](https://community.harness.io/) - Ask questions, find answers, and help other users.

[Report and Track A Bug](https://community.harness.io/c/bugs/17) - Find a bug? Please report in our forum under Drone Bugs. Please provide screenshots and steps to reproduce. 

[Events](https://www.meetup.com/harness/) - Keep up to date with Drone events and check out previous events [here](https://www.youtube.com/watch?v=Oq34ImUGcHA&list=PLXsYHFsLmqf3zwelQDAKoVNmLeqcVsD9o).
