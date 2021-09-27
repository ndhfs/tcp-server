package tcp

import "errors"

var (
	ErrServerIsBusy = errors.New("ServerIsBusy")
	ErrInvalidPackage =  errors.New("ErrInvalidPackage")
	ErrUndefinedConn =  errors.New("ErrUndefinedConn")
	ErrHandlerNotRegistered =  errors.New("ErrHandlerNotRegistered")
)
