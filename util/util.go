package util

func LenString(length int, buf string) string {
	data := make([]byte, length)
	if length > len(buf) {
		length = len(buf)
	}
	copy(data, buf[:length])
	return string(data)
}
