// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"os"
	"time"
)

const (
	JacocoTool = "jacoco"
	JunitTool  = "junit"
	NunitTool  = "nunit"
	TestNgTool = "testng"

	PipeLineIdEnvVar  = "HARNESS_PIPELINE_ID"
	BuildNumberEnvVar = "HARNESS_BUILD_ID"
)

type ResultBasicInfo struct {
	PipelineId string
	BuildId    string
	Name       string
	Type       string
	Status     string
}

type DbCredentials struct {
	InfluxDBURL   string
	InfluxDBToken string
	Organization  string
	Bucket        string
}

func PersistToInfluxDb(dbUrl, dbToken, dbOrganisation, dbBucket, measurementName string,
	tagsMap map[string]string, fieldsMap map[string]interface{}) error {

	client := influxdb2.NewClient(dbUrl, dbToken)
	defer client.Close()

	writeAPI := client.WriteAPIBlocking(dbOrganisation, dbBucket)
	point := influxdb2.NewPoint(
		measurementName,
		tagsMap,
		fieldsMap,
		time.Now())
	err := writeAPI.WritePoint(context.Background(), point)
	if err != nil {
		fmt.Println("Error writing point: ", err)
		return err
	}
	fmt.Println("Data persisted successfully to InfluxDB.")
	return nil
}

func ToStructFromJsonString[T any](jsonStr string) (T, error) {
	var v T
	err := json.Unmarshal([]byte(jsonStr), &v)
	return v, err
}

func ToJsonStringFromStruct[T any](v T) (string, error) {
	jsonBytes, err := json.Marshal(v)

	if err == nil {
		return string(jsonBytes), nil
	}

	return "", err
}

func GetPipelineInfo() (string, string, error) {
	pipelineId := os.Getenv(PipeLineIdEnvVar)
	buildNumber := os.Getenv(BuildNumberEnvVar)

	if pipelineId == "" || buildNumber == "" {
		return "", "", fmt.Errorf("PipelineId or BuildNumber not found in the environment")
	}

	return pipelineId, buildNumber, nil
}
