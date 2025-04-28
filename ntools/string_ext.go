package ntools

type NString struct {
	S string
}

func (ns NString) CutStr(start, end int) string {
	return ns.S[start:end]
}

// CutString ...
func (ns NString) CutString(length int) string {
	if len(ns.S) > length {
		if length > 6 {
			resultStr := ns.S[0 : length-3]
			return resultStr + "..."
		}
		return ns.S[0:length]
	}
	return ns.S
}
