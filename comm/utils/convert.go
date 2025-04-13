package utils

import "claimask/comm/constant"

func DogeToElon(amount float64) int64 {
	return int64(amount * constant.DOGE_TO_ELON)
}

func ElonToDoge(value int64) float64 {
	return float64(value) * constant.ELON_TO_DOGE
}
