package flow

import (
	"fmt"
	"math/big"
	"strings"
)

func formatAmount(raw string, decimals int, precision int) string {
	if raw == "" {
		return "0"
	}

	value := new(big.Int)
	if _, ok := value.SetString(raw, 10); !ok {
		return raw
	}

	if decimals == 0 {
		return value.String()
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	intPart := new(big.Int).Div(value, divisor)
	fracPart := new(big.Int).Mod(value, divisor)

	fracString := fmt.Sprintf("%0*s", decimals, fracPart.String())
	fracString = strings.TrimRight(fracString, "0")
	if precision > 0 && len(fracString) > precision {
		fracString = fracString[:precision]
	}
	if fracString == "" {
		return intPart.String()
	}

	return intPart.String() + "." + fracString
}
