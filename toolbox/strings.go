package toolbox

func TruncateStr(s string, maxlen int) string {
	if len(s) <= maxlen {
		return s
	}
	return s[:maxlen] + "..."
}
