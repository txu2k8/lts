package logger

import (
	"encoding/json"
	"fmt"
	"stress/config"
	"strings"
	"sync"
	"unicode"

	"github.com/cheggaaa/pb"
	"github.com/minio/mc/pkg/probe"
	"github.com/minio/pkg/console"
)

// CauseMessage container for golang error messages
type CauseMessage struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}

// ErrorMessage container for error messages
type ErrorMessage struct {
	Message   string             `json:"message"`
	Cause     CauseMessage       `json:"cause"`
	Type      string             `json:"type"`
	CallTrace []probe.TracePoint `json:"trace,omitempty"`
	SysInfo   map[string]string  `json:"sysinfo"`
}

var PrintMu sync.Mutex

func PrintInfo(data ...interface{}) {
	PrintMu.Lock()
	defer PrintMu.Unlock()
	w, _ := pb.GetTerminalWidth()
	if w > 0 {
		fmt.Print("\r", strings.Repeat(" ", w), "\r")
	} else {
		data = append(data, "\n")
	}
	console.Info(data...)
}

func PrintError(data ...interface{}) {
	PrintMu.Lock()
	defer PrintMu.Unlock()
	w, _ := pb.GetTerminalWidth()
	if w > 0 {
		fmt.Print("\r", strings.Repeat(" ", w), "\r")
	} else {
		data = append(data, "\n")
	}
	console.Errorln(data...)
}

// FatalIf wrapper function which takes error and selectively prints stack frames if available on debug
func FatalIf(err *probe.Error, msg string, data ...interface{}) {
	if err == nil {
		return
	}
	Fatal(err, msg, data...)
}

func Fatal(err *probe.Error, msg string, data ...interface{}) {
	if config.GlobalJSON {
		errorMsg := ErrorMessage{
			Message: msg,
			Type:    "fatal",
			Cause: CauseMessage{
				Message: err.ToGoError().Error(),
				Error:   err.ToGoError(),
			},
			SysInfo: err.SysInfo,
		}
		if config.GlobalDebug {
			errorMsg.CallTrace = err.CallTrace
		}
		json, e := json.MarshalIndent(struct {
			Status string       `json:"status"`
			Error  ErrorMessage `json:"error"`
		}{
			Status: "error",
			Error:  errorMsg,
		}, "", " ")
		if e != nil {
			console.Fatalln(probe.NewError(e))
		}
		console.Infoln(string(json))
		console.Fatalln()
	}

	msg = fmt.Sprintf(msg, data...)
	errmsg := err.String()
	if !config.GlobalDebug {
		errmsg = err.ToGoError().Error()
	}

	// Remove unnecessary leading spaces in generic/detailed error messages
	msg = strings.TrimSpace(msg)
	errmsg = strings.TrimSpace(errmsg)

	// Add punctuations when needed
	if len(errmsg) > 0 && len(msg) > 0 {
		if msg[len(msg)-1] != ':' && msg[len(msg)-1] != '.' {
			// The detailed error message starts with a capital letter,
			// we should then add '.', otherwise add ':'.
			if unicode.IsUpper(rune(errmsg[0])) {
				msg += "."
			} else {
				msg += ":"
			}
		}
		// Add '.' to the detail error if not found
		if errmsg[len(errmsg)-1] != '.' {
			errmsg += "."
		}
	}
	fmt.Println("")
	console.Fatalln(fmt.Sprintf("%s %s", msg, errmsg))
}

// ErrorIf synonymous with fatalIf but doesn't exit on error != nil
func ErrorIf(err *probe.Error, msg string, data ...interface{}) {
	if err == nil {
		return
	}
	if config.GlobalJSON {
		errorMsg := ErrorMessage{
			Message: fmt.Sprintf(msg, data...),
			Type:    "error",
			Cause: CauseMessage{
				Message: err.ToGoError().Error(),
				Error:   err.ToGoError(),
			},
			SysInfo: err.SysInfo,
		}
		if config.GlobalDebug {
			errorMsg.CallTrace = err.CallTrace
		}
		json, e := json.MarshalIndent(struct {
			Status string       `json:"status"`
			Error  ErrorMessage `json:"error"`
		}{
			Status: "error",
			Error:  errorMsg,
		}, "", " ")
		if e != nil {
			console.Fatalln(probe.NewError(e))
		}
		console.Infoln(string(json))
		return
	}
	msg = fmt.Sprintf(msg, data...)
	if !config.GlobalDebug {
		console.Errorln(fmt.Sprintf("%s %s", msg, err.ToGoError()))
		return
	}
	fmt.Println("")
	console.Errorln(fmt.Sprintf("%s %s", msg, err))
}
