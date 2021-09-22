package server

import "errors"

var (
	ErrServerIsBusy = errors.New("ServerIsBusy")
	ErrInvalidPackage =  errors.New("ErrInvalidPackage")
)
