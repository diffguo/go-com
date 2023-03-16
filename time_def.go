package gocom

import (
	"fmt"
	"github.com/diffguo/gocom/goroutineid"
	"github.com/diffguo/gocom/log"
	"runtime"
	"time"
)

const (
	StdDateFormat                  = "2006-01-02"
	StdTimeFormat                  = "15:04:05"
	StdDateTimeFormat              = "2006-01-02 15:04:05"
	StdTimeWithMsec                = "15:04:05.999999999"
	StdDateTimeWithMsec            = "2006-01-02 15:04:05.999999999"
	StdTimeWithMsecAndTimeZone     = "15:04:05.999999999Z07:00"
	StdDateTimeWithMsecAndTimeZone = "2006-01-02 15:04:05.999999999Z07:00"
)

// 函数运行计时器，用法，在函数开始的地方添加： defer TimeCounter()()
func TimeCounter() func() {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])

	start := time.Now()
	if log.GLog != nil {
		log.InfoF("[%d] enter func: %s", goroutineid.GetGoID(), f.Name())
	} else {
		fmt.Printf("[%s][%d] enter func: %s\n", time.Now().Format(StdTimeWithMsec), goroutineid.GetGoID(), f.Name())
	}

	return func() {
		if log.GLog != nil {
			log.InfoF("[%d] exit func: %s after: %s \n", goroutineid.GetGoID(), f.Name(), time.Since(start))
		} else {
			fmt.Printf("[%s][%d] exit func: %s after: %s \n", time.Now().Format(StdTimeWithMsec), goroutineid.GetGoID(), f.Name(), time.Since(start))
		}
	}
}
