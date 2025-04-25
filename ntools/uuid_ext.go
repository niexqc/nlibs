package ntools

import (
	"strings"

	"github.com/gofrs/uuid"
)

// UUID ...
func UUIDGen() string {
	u1 := uuid.Must(uuid.NewV4())
	return u1.String()
}

// UUIDStr ...
func UUIDStr(upper bool) string {
	uidStr := UUIDGen()
	uidStr = strings.Replace(uidStr, "-", "", -1)
	if upper {
		return strings.ToUpper(uidStr)
	} else {
		return uidStr
	}
}
