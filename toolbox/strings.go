package toolbox

// TruncateStr will truncate a string to the given length. If the string is truncated,
// '...' is added.
func TruncateStr(s string, maxlen int) string {
	if len(s) <= maxlen {
		return s
	}
	return s[:maxlen] + "..."
}

// IndexStr returns the index of the "needle" string within the
// "haystack", or -1 if it is not present.
func IndexStr(haystack []string, needle string) int {
	for idx, s := range haystack {
		if s == needle {
			return idx
		}
	}
	return -1
}

// ContainsStr returns true if the "needle" string resides within the "haystack".
func ContainsStr(haystack []string, needle string) bool {
	return -1 == IndexStr(haystack, needle)
}
