package plugin

import (
	"errors"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/harness-community/parse-test-reports/gojunit"
	"github.com/mattn/go-zglob"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type JunitAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
	DbCredentials
}

type TestStats struct {
	TestCount    int
	FailCount    int
	PassCount    int
	SkippedCount int
	ErrorCount   int
}

func GetNewJunitAggregator(
	reportsDir, reportsName, includes, dbUrl, dbToken, dbOrg, dbBucket string) *JunitAggregator {
	return &JunitAggregator{
		ReportsDir:  reportsDir,
		ReportsName: reportsName,
		Includes:    includes,
		DbCredentials: DbCredentials{
			InfluxDBURL:   dbUrl,
			InfluxDBToken: dbToken,
			Organization:  dbOrg,
			Bucket:        dbBucket,
		},
	}
}

func (j *JunitAggregator) Aggregate(groupName string) error {
	logrus.Println("JunitAggregator Aggregator Aggregate")

	reportsRootDir := j.ReportsDir
	includes := j.Includes
	var xmlReportFiles []string
	patterns := strings.Split(includes, ",")

	for _, pattern := range patterns {
		tmpReportDir := os.DirFS(reportsRootDir)
		relPattern := strings.TrimPrefix(pattern, reportsRootDir+"/")

		filesList, err := doublestar.Glob(tmpReportDir, relPattern)
		if err != nil {
			logrus.Println("Include patterns not found ", err.Error())
			return err
		}
		xmlReportFiles = append(xmlReportFiles, filesList...)
	}

	for i, tmpXmlReportFile := range xmlReportFiles {
		xmlReportFiles[i] = filepath.Join(reportsRootDir, tmpXmlReportFile)
	}

	totalAggregate, err := ParseTests(xmlReportFiles, logrus.New())
	if err != nil {
		logrus.Println("error: ", err)
	}

	pipelineId, buildNumber, err := GetPipelineInfo()
	if err != nil {
		logrus.Println("Error getting pipeline info: ", err.Error())
		return err
	}

	tagsMap, fieldsMap := GetJunitDataMaps(pipelineId, buildNumber, totalAggregate)
	err = ShowJunitStats(tagsMap, fieldsMap)
	if err != nil {
		logrus.Println("Error showing build stats: ", err.Error())
		return err
	}

	err = ExportJunitOutputVars(tagsMap, fieldsMap)
	if err != nil {
		logrus.Println("Error exporting Junit output vars: ", err.Error())
		return err
	}

	if j.DbCredentials.InfluxDBURL != "" && j.DbCredentials.InfluxDBToken != "" &&
		j.DbCredentials.Organization != "" && j.DbCredentials.Bucket != "" {
		err = PersistToInfluxDb(j.DbCredentials.InfluxDBURL, j.DbCredentials.InfluxDBToken,
			j.DbCredentials.Organization, j.DbCredentials.Bucket, JunitTool, groupName, tagsMap, fieldsMap)
		if err != nil {
			logrus.Println("Error persisting data to InfluxDB: ", err.Error())
			return err
		}
	}

	return err
}

func GetJunitDataMaps(pipelineId, buildNumber string, aggregateData TestStats) (map[string]string, map[string]interface{}) {
	tags := map[string]string{
		"pipelineId": pipelineId,
		"buildId":    buildNumber,
	}

	fields := map[string]interface{}{
		"total_tests":   aggregateData.TestCount,
		"failed_tests":  aggregateData.FailCount,
		"passed_tests":  aggregateData.PassCount,
		"skipped_tests": aggregateData.SkippedCount,
		"errors_count":  aggregateData.ErrorCount,
	}

	return tags, fields
}

func ExportJunitOutputVars(tags map[string]string, fields map[string]interface{}) error {
	outputVarsMap := map[string]interface{}{
		"TOTAL_CASES":   fields["total_tests"],
		"TOTAL_PASSED":  fields["passed_tests"],
		"TOTAL_FAILED":  fields["failed_tests"],
		"TOTAL_SKIPPED": fields["skipped_tests"],
		"TOTAL_ERRORS":  fields["errors_count"],
	}
	for key, value := range outputVarsMap {
		err := WriteToEnvVariable(key, fmt.Sprintf("%v", value))
		if err != nil {
			logrus.Errorf("Error writing %s to env variable: %v", key, err)
			return err
		}
	}
	return nil
}

func ShowJunitStats(tags map[string]string, fields map[string]interface{}) error {
	border := "============================================="
	separator := "---------------------------------------------"

	table := []string{
		border,
		"  JUnit Test Run Summary",
		border,
		fmt.Sprintf("  Pipeline ID: %-40s", tags["pipelineId"]),
		fmt.Sprintf("  Build ID: %-40s", tags["buildId"]),
		border,
		fmt.Sprintf("| %-22s | %-10s |", "Test Category", "Count            "),
		separator,
		fmt.Sprintf("| 📁 Total Cases      | %10.2f          |", float64(fields["total_tests"].(int))),
		fmt.Sprintf("| ✅ Total Passed     | %10.2f          |", float64(fields["passed_tests"].(int))),
		fmt.Sprintf("| ❌ Total Failed     | %10.2f          |", float64(fields["failed_tests"].(int))),
		fmt.Sprintf("| ⏸️ Total Skipped    | %10.2f          |", float64(fields["skipped_tests"].(int))),
		fmt.Sprintf("| 🛑 Total Errors     | %10.2f          |", float64(fields["errors_count"].(int))),
		border,
	}

	fmt.Println(strings.Join(table, "\n"))
	return nil
}

func CompareJunitResults(tool string, args Args) (string, error) {
	var resultStr string
	currentPipelineId, currentBuildNumber, err := GetPipelineInfo()
	if err != nil {
		fmt.Println("CompareResults Error getting pipeline info: ", err)
		return resultStr, err
	}

	previousBuildId, err := GetPreviousBuildId(tool, args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket, currentPipelineId, args.GroupName, currentBuildNumber, args)
	if err != nil {
		fmt.Println("CompareResults Error getting previous build id: ", err)
		return resultStr, err
	}

	resultStr, err = GetComparedDifferences(tool, args.DbUrl, args.DbToken, args.DbOrg, args.DbBucket,
		currentPipelineId, args.GroupName, currentBuildNumber, strconv.Itoa(previousBuildId))
	if err != nil {
		fmt.Println("CompareResults Error getting compared differences: ", err)
		return resultStr, err
	}
	return resultStr, nil
}

func ParseTests(paths []string, log *logrus.Logger) (TestStats, error) {
	files := getFiles(paths, log)
	stats := TestStats{}

	if len(files) == 0 {
		log.Errorln("could not find any files matching the provided report path")
		return stats, nil
	}

	for _, file := range files {
		suites, err := gojunit.IngestFile(file)
		if err != nil {
			log.WithError(err).WithField("file", file).Errorln("could not parse file")
			continue
		}
		fileStats := TestStats{}
		for _, suite := range suites {
			for _, test := range suite.Tests {
				fileStats.TestCount++
				switch test.Result.Status {
				case "passed":
					fileStats.PassCount++
				case "failed":
					fileStats.FailCount++
				case "skipped":
					fileStats.SkippedCount++
				case "error":
					fileStats.ErrorCount++
				}
			}
		}

		// Aggregate stats
		stats.TestCount += fileStats.TestCount
		stats.PassCount += fileStats.PassCount
		stats.FailCount += fileStats.FailCount
		stats.SkippedCount += fileStats.SkippedCount
		stats.ErrorCount += fileStats.ErrorCount
	}

	if stats.FailCount > 0 || stats.ErrorCount > 0 {
		return stats, errors.New("failed tests and errors found")
	}
	return stats, nil
}

func getFiles(paths []string, log *logrus.Logger) []string {
	var files []string
	for _, p := range paths {
		path, err := expandTilde(p)
		if err != nil {
			log.WithError(err).WithField("path", p).Errorln("error expanding path")
			continue
		}
		matches, err := zglob.Glob(path)
		if err != nil {
			log.WithError(err).WithField("path", path).Errorln("error resolving path regex")
			continue
		}

		files = append(files, matches...)
	}
	return uniqueItems(files)
}

func uniqueItems(items []string) []string {
	var result []string
	set := make(map[string]bool)
	for _, item := range items {
		if !set[item] {
			result = append(result, item)
			set[item] = true
		}
	}
	return result
}

func expandTilde(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if path[0] != '~' {
		return path, nil
	}
	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", errors.New("cannot expand user-specific home dir")
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, path[1:]), nil
}

func LoadYAML(source string) (map[string]interface{}, error) {
	log := logrus.New()
	log.Infoln("Loading YAML from source:", source)

	var data []byte
	var err error

	if isURL(source) {
		resp, err := http.Get(source)
		if err != nil {
			log.WithError(err).Errorln("Failed to fetch YAML from URL")
			return nil, err
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			log.WithError(err).Errorln("Failed to read YAML data from URL")
			return nil, err
		}
	} else {
		data, err = os.ReadFile(source)
		if err != nil {
			log.WithError(err).Errorln("Failed to read local YAML file")
			return nil, err
		}
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		log.WithError(err).Errorln("Failed to parse YAML")
		return nil, err
	}

	log.Infoln("Successfully loaded and parsed YAML")
	return result, nil
}

func isURL(source string) bool {
	return strings.HasPrefix(source, "http")
}

func ParseTestsWithQuarantine(paths []string, quarantineList map[string]interface{}, log *logrus.Logger) (TestStats, error) {
	files := getFiles(paths, log)
	stats := TestStats{}
	nonQuarantinedFailures := 0
	expiredTests := 0

	if len(files) == 0 {
		log.Errorln("could not find any files matching the provided report path")
		return stats, nil
	}

	log.Infoln("Starting to parse tests with quarantine list")

	for _, file := range files {
		suites, err := gojunit.IngestFile(file)
		if err != nil {
			log.WithError(err).WithField("file", file).Errorln("could not parse file")
			continue
		}
		fileStats := TestStats{}
		for _, suite := range suites {
			for _, test := range suite.Tests {
				fileStats.TestCount++
				testIdentifier := test.Classname + "." + test.Name
				switch test.Result.Status {
				case "passed":
					fileStats.PassCount++
				case "failed":
					if !isQuarantined(testIdentifier, quarantineList, log) {
						log.Infoln("Not Quarantined test failed:", testIdentifier)
						nonQuarantinedFailures++
					} else if isExpired(testIdentifier, quarantineList, log) {
						log.Infoln("Quarantined test expired:", testIdentifier)
						expiredTests++
					}
					fileStats.FailCount++
				case "skipped":
					fileStats.SkippedCount++
				case "error":
					fileStats.ErrorCount++
				}
			}
		}

		stats.TestCount += fileStats.TestCount
		stats.PassCount += fileStats.PassCount
		stats.FailCount += fileStats.FailCount
		stats.SkippedCount += fileStats.SkippedCount
		stats.ErrorCount += fileStats.ErrorCount
	}

	if nonQuarantinedFailures > 0 || expiredTests > 0 {
		// Construct the error message by concatenating string values
		errorMessage := "Non-quarantined failures: " + strconv.Itoa(nonQuarantinedFailures) +
			", Expired tests: " + strconv.Itoa(expiredTests) + " found"
		return stats, errors.New(errorMessage)
	}

	return stats, nil
}

func isQuarantined(testIdentifier string, quarantineList map[string]interface{}, log *logrus.Logger) bool {
	log.Infoln("Checking if test is quarantined:", testIdentifier)
	tests, ok := quarantineList["quarantine_tests"].([]interface{})
	if !ok {
		log.Warnln("Quarantine list invalid or missing 'quarantine_tests'")
		return false
	}
	for _, test := range tests {
		if testMap, ok := test.(map[interface{}]interface{}); ok {
			if quarantinedIdentifier, found := matchTestIdentifier(testMap, testIdentifier, log); found {
				log.Infoln("Test is quarantined:", quarantinedIdentifier)
				return true
			}
		}
	}
	log.Infoln("Test is not quarantined:", testIdentifier)
	return false
}

func isExpired(testIdentifier string, quarantineList map[string]interface{}, log *logrus.Logger) bool {
	tests, ok := quarantineList["quarantine_tests"].([]interface{})
	if !ok {
		log.Warnln("Quarantine list invalid or missing 'quarantine_tests'")
		return false
	}
	for _, test := range tests {
		if testMap, ok := test.(map[interface{}]interface{}); ok {
			if quarantinedIdentifier, found := matchTestIdentifier(testMap, testIdentifier, log); found {
				startDate, startOk := testMap["start_date"].(string)
				endDate, endOk := testMap["end_date"].(string)

				if startOk && endOk {
					currentDate := time.Now()

					startTime, err := time.Parse("2006-01-02", startDate)
					if err != nil {
						log.WithError(err).Warnln("Failed to parse start_date")
						continue
					}

					endTime, err := time.Parse("2006-01-02", endDate)
					if err != nil {
						log.WithError(err).Warnln("Failed to parse end_date")
						continue
					}

					if currentDate.Before(startTime) || currentDate.After(endTime) {
						log.WithFields(logrus.Fields{
							"test":        quarantinedIdentifier,
							"currentDate": currentDate,
							"startDate":   startTime,
							"endDate":     endTime,
						}).Infoln("Current Date lies outside start_date and end_date.")
						return true
					}
				}
			}
		}
	}

	log.Infoln("Test has no expiration set:", testIdentifier)
	return false
}

func matchTestIdentifier(testMap map[interface{}]interface{}, identifier string, log *logrus.Logger) (string, bool) {
	quarantinedClassname, classnameOk := testMap["classname"].(string)
	quarantinedName, nameOk := testMap["name"].(string)

	if classnameOk && nameOk {
		quarantinedIdentifier := quarantinedClassname + "." + quarantinedName
		if quarantinedIdentifier == identifier {
			log.Infoln("Test", identifier, "is quarantined")
			return quarantinedIdentifier, true
		}
	}
	return "", false
}
