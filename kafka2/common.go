package kafka2

import (
	"errors"
	"fmt"
	"github.com/4books-sparta/utils/logging"
	"os"
	"runtime/debug"
)

type Stoppable interface {
	Stop() error
}

func PanicHandler(rep logging.ErrorReporter, clientId string, k Stoppable) {
	r := recover()
	if r == nil {
		return // no panic underway
	}

	rep.Report(errors.New("panic-in-consumer"), "e", fmt.Sprintf("%v", r), "cid", clientId)

	fmt.Printf("PanicHandler invoked because %v\n", r)

	// print debug stack
	debug.PrintStack()
	if k != nil {
		_ = k.Stop()
	}

	os.Exit(1)
}
