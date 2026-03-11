package filelog

import (
	"io"
	"os"
	"strings"

	"github.com/eleanorhealth/go-common/pkg/errs"
)

type Logger interface {
	Log(msg string) error
}

type Nop struct{}

var _ Logger = Nop{}

func (Nop) Log(msg string) error {
	return nil
}

type FileLogger struct {
	Path string
}

var _ Logger = FileLogger{}

func (f FileLogger) Log(msg string) error {
	file, err := os.OpenFile(f.Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm|os.ModeAppend)
	if err != nil {
		return errs.Wrap(err, "unable to open file")
	}

	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = errs.Wrap(e, "unable to close file")
		}
	}()
	_, err = io.Copy(file, strings.NewReader(msg+"\n"))
	if err != nil {
		return errs.Wrap(err, "unable to write to file")
	}

	return nil
}
