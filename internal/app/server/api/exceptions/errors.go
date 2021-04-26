package exceptions

import "errors"

var (
	DatabaseDumpFailed        = errors.New("could not create database dump")
	DumpFileNotFoundInRequest = errors.New("no file field found in form/multipart section of request")
	DumpSaveFailed            = errors.New("can not save specified file")

	IncorrectAuthData     = errors.New("incorrect email or password")
	IncorrectRefreshToken = errors.New("incorrect refresh token")
)