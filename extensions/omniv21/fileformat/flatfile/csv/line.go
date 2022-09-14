package csv

type line struct {
	lineNum int // 1-based
	record  []string
	raw     string
}
