package testing

import (
	"log"

	"github.com/KompiTech/fabric-cc-core/v2/internal/testing"
)

// InitializeCouchDBContainer prepares CouchDB running in docker for use by this mock. It should be called only once before test suite, because it is quite time intensive
func InitializeCouchDBContainer() {
	if !testing.IsCouchRunning() {
		log.Print("CouchDB container is not running, removing it...")
		testing.RemoveCouch()
		log.Print("Creating new CouchDB container")
		testing.RunCouch()
	}
	testing.WaitForCouch()
	testing.InitSysDBs()
}
