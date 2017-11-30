package notbearclient

import (
	"fmt"
	"net/url"
)

// type ErrTimeout struct {
// 	URL         *url.URL
// 	Times       int
// 	OriginError error
// }
//
// func NewErrTimeout(url *url.URL, times int, err error) *ErrTimeout {
// 	return &ErrTimeout{
// 		URL:         url,
// 		Times:       times,
// 		OriginError: err,
// 	}
// }
//
// func (et *ErrTimeout) Error() string {
// 	return fmt.Sprintf("The %d time retry of %s has failed (due to %s)", et.Times, et.URL, et.OriginError)
// }

type ErrTimeout struct {
	URL         *url.URL
	OriginError error
}

func NewErrTimeout(url *url.URL, err error) *ErrTimeout {
	return &ErrTimeout{
		URL:         url,
		OriginError: err,
	}
}

func (et *ErrTimeout) Error() string {
	return fmt.Sprintf("Timeout of accessing %s (due to %s)", et.URL, et.OriginError)
}

// type ErrFailed struct {
// 	URL         *url.URL
// 	OriginError error
// }
//
// func NewErrFailed(url *url.URL, err error) *ErrFailed {
// 	return &ErrFailed{
// 		URL:         url,
// 		OriginError: err,
// 	}
// }
//
// func (ef *ErrFailed) Error() string {
// 	return fmt.Sprintf("Failed to access %s (due to %s)", ef.URL, ef.OriginError)
// }

type ErrNetwork struct {
	URL         *url.URL
	OriginError error
}

func NewErrNetwork(url *url.URL, err error) *ErrNetwork {
	return &ErrNetwork{
		URL:         url,
		OriginError: err,
	}
}

func (en *ErrNetwork) Error() string {
	return fmt.Sprintf("Network error while accessing %s (due to %s)", en.URL, en.OriginError)
}

type ErrOther struct {
	URL         *url.URL
	OriginError error
}

func NewErrOther(url *url.URL, err error) *ErrOther {
	return &ErrOther{
		URL:         url,
		OriginError: err,
	}
}

func (eo *ErrOther) Error() string {
	return fmt.Sprintf("Other error while accessing %s (due to %s)", eo.URL, eo.OriginError)
}
