package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Findings struct {
	Finding []StaticJson `json:"findings"`
}

type StaticJson struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Location struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
	} `json:"location"`
	End struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
	} `json:"end"`
	Message string `json:"message"`
}

// CodeClimateIssue is a subset of the Code Climate spec.
// https://github.com/codeclimate/platform/blob/master/spec/analyzers/SPEC.md#data-types
// It is just enough to support GitLab CI Code Quality.
// https://docs.gitlab.com/ee/user/project/merge_requests/code_quality.html

type CodeClimateIssue struct {
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Fingerprint string `json:"fingerprint"`
	Location    struct {
		Path  string `json:"path"`
		Lines struct {
			Begin int `json:"begin"`
		} `json:"lines"`
	} `json:"location"`
}

func main() {

	jsonFile, err := os.Open("staticcheck.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	fscanner := bufio.NewScanner(jsonFile)

	// fix staticcheck json output with bad structure
	byteValue := "{ \"findings\": [ "
	for fscanner.Scan() {
		byteValue = fmt.Sprintf("%v%v%v", byteValue, fscanner.Text(), ",")
	}
	byteValue = strings.TrimRight(byteValue, ",")
	paddedJson := fmt.Sprintf("%v%v", byteValue, ("]}"))

	// turn json file into findings struct
	findings := &Findings{}
	err = json.Unmarshal([]byte(paddedJson), findings)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	currentDir, _ := os.Getwd()

	// convert staticcheck output to code climate issues
	codeClimateIssues := make([]CodeClimateIssue, 0, len(findings.Finding))
	for _, finding := range findings.Finding {
		codeClimateIssue := CodeClimateIssue{}
		codeClimateIssue.Description = finding.Code + ": " + finding.Message
		codeClimateIssue.Location.Path = strings.TrimPrefix(finding.Location.File, currentDir+"/")
		codeClimateIssue.Location.Lines.Begin = finding.Location.Line

		if finding.Severity == "error" {
			codeClimateIssue.Severity = "critical"
		} else {
			codeClimateIssue.Severity = "warning"
		}

		// Get issue fingerprint
		h := md5.New()
		_, _ = fmt.Fprintf(h, "%s%s%s", codeClimateIssue.Location.Path, codeClimateIssue.Description, string(codeClimateIssue.Location.Lines.Begin))
		md5sum := fmt.Sprintf("%X", h.Sum(nil))
		codeClimateIssue.Fingerprint = md5sum

		codeClimateIssues = append(codeClimateIssues, codeClimateIssue)

	}

	// outputting code quality report
	outputJSON, err := json.MarshalIndent(codeClimateIssues, "", "    ")
	if err != nil {
		log.Fatal("Error marshalling output: ", err)
	}

	fmt.Printf("code quality report:\n%s", string(outputJSON))

	err = os.WriteFile("gl-code-quality-report.json", outputJSON, 0644)
	if err != nil {
		log.Fatal("Error during output: ", err)
	}
}
