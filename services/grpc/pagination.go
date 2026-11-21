package services

func fixPositiveNum(num, defaultValue int) int {
	if num < 1 {
		return defaultValue
	}
	return num
}

func fixPositiveNumMax(num, maxValue int) int {
	if num > maxValue {
		return maxValue
	}
	return num
}
