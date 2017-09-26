package pid

import (
	"errors"
	"os"
	"strconv"
	"syscall"
)

var (
	ErrProcessExists = errors.New("Process already exists.")
)

type PidFile struct {
	path string
	Pid  int
}

func New(path string) (pf *PidFile, err error) {
	pf = &PidFile{path, os.Getpid()}
	var f *os.File
	f, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil { // file exists
		f, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return
		}
		var pidstr [10]byte
		var n int
		// read pid
		n, err = f.Read(pidstr[:])
		if err != nil {
			return
		}
		var pid int
		pid, err = strconv.Atoi(string(pidstr[:n]))
		if err != nil {
			return
		}
		// find pid
		if err = syscall.Kill(pid, 0); err == nil {
			err = ErrProcessExists
			return
		}
		f.Truncate(int64(n))
	}
	defer f.Close()
	_, err = f.WriteString(strconv.Itoa(pf.Pid))
	return
}

func (pf *PidFile) Close() error {
	return os.Remove(pf.path)
}
