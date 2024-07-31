package common

import "strings"

// Concatenates symbol names
func JoinSymbolName(symbols []Symbol, separator string) string {
	var symbolsStr []string

	for _, symbol := range symbols {
		symbolsStr = append(symbolsStr, symbol.Name)
	}

	return strings.Join(symbolsStr, separator)
}
