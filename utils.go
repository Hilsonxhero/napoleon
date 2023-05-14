package napoleon

import (
	"fmt"
	"regexp"
	"runtime"
	"time"
)

func (n *Napoleon) LoadTime(start time.Time) {
	elapsed := time.Since(start)

	pc, _, _, _ := runtime.Caller(1)

	funcObj := runtime.FuncForPC(pc)

	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)

	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	n.InfoLog.Println(fmt.Sprintf("Load time : %s took %s", name, elapsed))
}
