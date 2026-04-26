package luhn

import "strconv"

func ValidInt(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func ValidString(number string) bool {
	num, err := strconv.Atoi(number)
	if err != nil {
		return false
	}
	return ValidInt(num)
}

func checksum(number int) int {
	var luhn int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}
		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
