package categoryguesser

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
)

type PreDefinedCategory struct {
	CategoryName string   `json:"CategoryName"`
	PayeeNames   []string `json:"PayeeNames"`
}

func loadPreDefinedCategories() ([]*PreDefinedCategory, error) {
	filePath := "pkg/category-guesser/pre_defined_categories.json"
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	var preDefinedCategories []*PreDefinedCategory

	if err := json.Unmarshal(jsonData, &preDefinedCategories); err != nil {
		return nil, err
	}

	for i := range preDefinedCategories {
		for j := range (*preDefinedCategories[i]).PayeeNames {
			payeeName := (*preDefinedCategories[i]).PayeeNames[j]
			(*preDefinedCategories[i]).PayeeNames[j] = preprocessText(payeeName)
		}
	}

	return preDefinedCategories, nil
}

func preprocessText(inputText string) string {
	processedText := strings.ToLower(inputText)
	processedText = strings.TrimSpace(processedText)
	re := regexp.MustCompile(`\s+`)
	processedText = re.ReplaceAllString(processedText, " ")
	re = regexp.MustCompile(`[^\w\s]`)
	processedText = re.ReplaceAllString(processedText, "")
	return processedText
}

func isPreDefinedCategory(transactionName string) (bool, string) {
	preDefinedCategories, err := loadPreDefinedCategories()
	if err != nil {
		return false, ""
	}

	for _, category := range preDefinedCategories {
		for _, payeeName := range (*category).PayeeNames {
			if strings.HasPrefix(payeeName, transactionName) {
				return true, (*category).CategoryName
			}
		}
	}

	splitedTransactionName := strings.Split(transactionName, " ")
	for _, category := range preDefinedCategories {
		for _, payeeName := range (*category).PayeeNames {
			curr := ""
			for idx := range splitedTransactionName {
				curr += splitedTransactionName[idx]

				if strings.HasPrefix(payeeName, curr) {
					return true, (*category).CategoryName
				}

				curr += " "
			}
		}
	}

	return false, ""
}

func guessCategoryAI(transactionName string) string {
	// TODO: Use AI
	return ""
}

func GuessCategory(transactionName string) (string, error) {
	transactionName = preprocessText(transactionName)

	if ok, categoryName := isPreDefinedCategory(transactionName); ok {
		return categoryName, nil
	}

	return guessCategoryAI(transactionName), nil
}
