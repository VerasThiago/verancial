package helper

import (
	"fmt"
	"os"
	"path"

	"github.com/verasthiago/verancial/app-integration-worker/pkg/types"
)

const PYTHON_SCRIPTS_PATH string = "../python-scripts/"

func GetPythonScriptsPath() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(pwd, PYTHON_SCRIPTS_PATH), nil
}

func GetFileNameFromAppReport(appReport types.AppReport) string {
	reportSize := len(appReport)
	if reportSize == 0 {
		return "empty"
	}

	if reportSize == 1 {
		return appReport[0][0]
	}

	firstDate := appReport[reportSize-1][0]
	lastDate := appReport[1][0]

	return fmt.Sprintf("%s_to_%s.csv", firstDate, lastDate)
}
