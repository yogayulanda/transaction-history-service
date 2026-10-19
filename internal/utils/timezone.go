package utils

import (
	"time"
	_ "time/tzdata"
)

func Jakarta() (*time.Location, error) {
	return time.LoadLocation("Asia/Jakarta")
}
