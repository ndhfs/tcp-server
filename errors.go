package tcp

import "errors"

var (
	ErrServerIsBusy = errors.New("ServerIsBusy")
	ErrInvalidPackage =  errors.New("ErrInvalidPackage")
	ErrHandlerNotRegistered =  errors.New("ErrHandlerNotRegistered")
)
