package cli

import (
	"errors"
	"s3stress/config"

	"github.com/minio/mc/pkg/probe"
)

type dummyErr error

var errDummy = func() *probe.Error {
	msg := ""
	return probe.NewError(dummyErr(errors.New(msg))).Untrace()
}

type invalidArgumentErr error

var errInvalidArgument = func() *probe.Error {
	msg := "Invalid arguments provided, please refer " + "`" + config.AppName + " <command> -h` for relevant documentation."
	return probe.NewError(invalidArgumentErr(errors.New(msg))).Untrace()
}
