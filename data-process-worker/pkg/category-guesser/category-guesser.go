package categoryguesser

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strings"
)

//go:embed pre_defined_categories.json
var predefinedCategoriesJSON []byte

type PreDefinedCategory struct {
	CategoryName string   `json:"CategoryName"`
	PayeeNames   []string `json:"PayeeNames"`
}

func loadPreDefinedCategories() ([]*PreDefinedCategory, error) {
	var preDefinedCategories []*PreDefinedCategory

	if err := json.Unmarshal(predefinedCategoriesJSON, &preDefinedCategories); err != nil {
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
		for _, payeeName := range category.PayeeNames {
			// Case 1: payee is longer — transaction is a prefix of the payee
			// e.g. transaction "FITNESS WORLD GEORGIA" matches payee "FITNESS WORLD GEORGIA RICHMOND BC"
			if strings.HasPrefix(payeeName, transactionName) {
				return true, category.CategoryName
			}
			// Case 2: single-word payee — transaction starts with payee at a word boundary
			// e.g. transaction "VIRGIN PLUS VERDUN QC" matches payee "VIRGIN"
			// Require word boundary (next char is space or end) to avoid "MARKET" matching "MARKETPLACE..."
			if !strings.Contains(payeeName, " ") && strings.HasPrefix(transactionName, payeeName) {
				rest := transactionName[len(payeeName):]
				if rest == "" || rest[0] == ' ' {
					return true, category.CategoryName
				}
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
