package util

import (
	log "github.com/xirtah/gopa-framework/core/logger/seelog"
)

// CheckProcessExists check if the pid is running
func CheckProcessExists(pid int) bool {
	log.Warn("process running check is not supported on Windows, please manually check with your working dir")
	return true
}
