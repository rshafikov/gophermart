package security

func LuhnAlgoPredicat(numeralID string) bool {
	sum := 0
	parity := len(numeralID) % 2

	for i, r := range numeralID {
		digit := int(r - '0')
		if i&1 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
