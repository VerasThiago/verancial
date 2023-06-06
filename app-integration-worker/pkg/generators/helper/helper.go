package helper

import (
	"os"
	"path"
)

const PYTHON_SCRIPTS_PATH string = "../python-scripts/"

func GetPythonScriptsPath() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(pwd, PYTHON_SCRIPTS_PATH), nil
}
