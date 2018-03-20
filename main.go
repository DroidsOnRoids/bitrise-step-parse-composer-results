package main

import (
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"path/filepath"
	"os"
	"regexp"
	"encoding/xml"
	"io/ioutil"
	"strings"
	"strconv"
)

// TestSuite ...
type TestSuite struct {
	Tests    int `xml:"tests,attr"`
	Failures int `xml:"failures,attr"`
	Errors   int `xml:"errors,attr"`
	Skipped  int `xml:"skipped,attr"`
}

func main() {
	allTestCount := 0
	passedTestCount := 0

	workdir := os.Getenv("results_dir")
	err := filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
		isMatchedReportFile, err := regexp.MatchString("reports/composer/.*/junit4-reports/.*\\.xml", path)
		if err != nil {
			return err
		}
		if isMatchedReportFile && info.Mode().IsRegular() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			var testSuite TestSuite
			err = xml.Unmarshal(data, &testSuite)
			if err != nil {
				return err
			}
			notPassed := testSuite.Errors + testSuite.Failures + testSuite.Skipped
			allTestCount += testSuite.Tests
			passedTestCount += testSuite.Tests - notPassed
		}
		return nil
	})

	if err != nil {
		log.Errorf("Composer results parse error: %s", err)
		os.Exit(1)
	}

	if err = exportEnvironmentWithEnvman("ALL_TEST_COUNT", strconv.Itoa(allTestCount)); err != nil {
		log.Errorf("Environment export error: %s", err)
		os.Exit(2)
	}

	if err = exportEnvironmentWithEnvman("PASSED_TEST_COUNT", strconv.Itoa(passedTestCount)); err != nil {
		log.Errorf("Environment export error: %s", err)
		os.Exit(3)
	}
}

func exportEnvironmentWithEnvman(keyStr, valueStr string) error {
	cmd := command.New("envman", "add", "--key", keyStr)
	cmd.SetStdin(strings.NewReader(valueStr))
	return cmd.Run()
}