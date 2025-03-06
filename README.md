# Test Result Aggregator Plugin

A plugin to aggregate test results from multiple testing tools and send them to an InfluxDB database.

# Usage

The following settings modify this plugin's behavior:

* `tool` - The testing framework to be used (e.g., `testng`, `nunit`, `junit`, `jacoco`).
* `group` - A logical grouping for test results.
* `reports_dir` - The directory containing test result files.
* `include_pattern` - A pattern to match test result files (e.g., `**/standard-TEST*.xml`).
* `influxdb_url` - The URL of the InfluxDB instance.
* `influxdb_token` - Authentication token for InfluxDB.
* `influxdb_org` - The organization in InfluxDB.
* `influxdb_bucket` - The InfluxDB bucket for storing test results.
* `compare_build_results` - Whether to compare results with previous builds. Specify true/false.
* `compare_build_id` - The specific build ID for comparison.

## Example `.drone.yml`

Below is an example `.drone.yml` that uses this plugin:

```yaml
kind: pipeline
name: default

steps:
- name: Run test result aggregator plugin
  image: plugins/test-result-aggregator-01-linux:latest
  settings:
    tool: junit
    group: suite_01
    reports_dir: /harness/junit
    include_pattern: "**/standard-TEST*.xml"
    influxdb_url: http://<your-url>:8086
    influxdb_token: <your-token>
    influxdb_org: hns
    influxdb_bucket: hns_test_bucket_02
    compare_build_results: true
    compare_build_id: 5
```

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t plugins/test-result-aggregator-01-linux -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm -e PLUGIN_TOOL=junit \
  -e PLUGIN_GROUP=suite_01 \
  -e PLUGIN_REPORTS_DIR=/harness/junit \
  -e PLUGIN_INCLUDE_PATTERN="**/standard-TEST*.xml" \
  -e PLUGIN_INFLUXDB_URL=http://<your-url>:8086 \
  -e PLUGIN_INFLUXDB_TOKEN=<your-token> \
  -e PLUGIN_INFLUXDB_ORG=hns \
  -e PLUGIN_INFLUXDB_BUCKET=hns_test_bucket_02 \
  -e PLUGIN_COMPARE_BUILD_RESULTS=true \
  -e PLUGIN_COMPARE_BUILD_ID=5 \
  plugins/test-result-aggregator-01-linux
```
