package util

import (
	"os"
	"path"

	log "github.com/xirtah/gopa-framework/core/logger/seelog"
)

var locked bool
var file string

// CheckInstanceLock make sure there is not a lock placed before check, and place a lock after check
func CheckInstanceLock(p string) {
	file = path.Join(p, ".lock")
	if FileExists(file) {
		log.Errorf("lock file:%s exists, if you only have one instance, please remove it", file)
		log.Flush()
		os.Exit(1)
	}
	FilePutContent(file, IntToString(os.Getpid()))
	log.Trace("lock placed,", file, " ,pid:", os.Getpid())
	locked = true
	log.Info("workspace: ", p)
}

// ClearInstanceLock remove the lock
func ClearInstanceLock() {
	if locked {
		os.Remove(path.Join(file))
	}
}
