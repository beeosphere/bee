package core

func ShortValue(value string) string {
	if len(value) > 8 {
		return ".." + value[len(value)-8:]
	}
	return value
}
