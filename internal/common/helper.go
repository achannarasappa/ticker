package common

import "strings"

// Concatenates symbol names
func JoinSymbolName(symbols []Symbol, separator string) string {
	symbolsStr := make([]string, len(symbols))

	for i, symbol := range symbols {
		symbolsStr[i] = symbol.Name
	}

	return strings.Join(symbolsStr, separator)
}
