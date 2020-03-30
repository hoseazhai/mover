package utils

import "math/rand"

var (
	letters = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGJKLMNOPQRSTUVWXYZ")
)

// RandID ... 生成指定长度的随机ID
func RandID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
