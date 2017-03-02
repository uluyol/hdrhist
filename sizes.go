package hdrhist

// This package doesn't actually do anything unsafe. We just
// need to import unsafe so that we can get the size of data
// types so we can return memory usage estimates to the user.
//
// Keep all uses of unsafe here so that we make sure unsafe
// is not imported in any of the other files.

import (
	"reflect"
	"time"
)

var histSize = int(reflect.TypeOf(Hist{}).Size())
var timeSize = int(reflect.TypeOf(time.Time{}).Size() + reflect.TypeOf(time.Location{}).Size())
