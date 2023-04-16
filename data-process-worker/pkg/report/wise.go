package report

import (
	"fmt"
)

type WiseReportProcessor struct{}

func (s WiseReportProcessor) ProcessReport(filePath string) error {
	fmt.Println("Processing Wise report:", filePath)
	return nil
}
