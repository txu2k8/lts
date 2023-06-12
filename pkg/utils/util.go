package utils

func Zfill(str string, width int) string {
	if len(str) >= width {
		return str
	}
	padding := make([]byte, width-len(str))
	for i := range padding {
		padding[i] = '0'
	}
	return string(padding) + str
}
