package plugin

import (
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

type JacocoAggregator struct {
	ReportsDir  string
	ReportsName string
	Includes    string
}

func (j *JacocoAggregator) Aggregate() error {
	fmt.Println("JacocoAggregator.Aggregate == ")

	var xmlReportFiles []string

	reportsRootDir := j.ReportsDir
	patterns := strings.Split(j.Includes, ",")

	fmt.Println("Include patterns: len(patterns) ", len(patterns))

	for _, pattern := range patterns {
		fmt.Println("pattern ==  ", pattern)
		fmt.Println("reportsRootDir ==  ", reportsRootDir)
		tmpReportDir := os.DirFS(reportsRootDir)
		relPattern := strings.TrimPrefix(pattern, reportsRootDir+"/")
		filesList, err := doublestar.Glob(tmpReportDir, relPattern)
		if err != nil {
			logrus.Println("Include patterns not found ", err.Error())
		}
		xmlReportFiles = append(xmlReportFiles, filesList...)
	}

	for _, xmlReportFile := range xmlReportFiles {
		fmt.Println("Processing file: ", xmlReportFile)
	}

	return nil
}
