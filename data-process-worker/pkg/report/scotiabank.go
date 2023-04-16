package report

import (
	"fmt"
)

type ScotiaBankReportProcessor struct{}

func (s ScotiaBankReportProcessor) ProcessReport(filePath string) error {
	fmt.Println("Processing ScotiaBank report:", filePath)
	return nil
}
