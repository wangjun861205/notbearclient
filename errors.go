package notbearclient

import (
	"fmt"
	"net/url"
)

type ErrTimeout struct {
	URL         *url.URL
	Times       int
	OriginError error
}

func NewErrTimeout(url *url.URL, times int, err error) *ErrTimeout {
	return &ErrTimeout{
		URL:         url,
		Times:       times,
		OriginError: err,
	}
}

func (et *ErrTimeout) Error() string {
	return fmt.Sprintf("The %d time retry of %s has failed (due to %s)", et.Times, et.URL, et.OriginError)
}

type ErrFailed struct {
	URL         *url.URL
	OriginError error
}

func NewErrFailed(url *url.URL, err error) *ErrFailed {
	return &ErrFailed{
		URL:         url,
		OriginError: err,
	}
}

func (ef *ErrFailed) Error() string {
	return fmt.Sprintf("Failed to access %s (due to %s)", ef.URL, ef.OriginError)
}
