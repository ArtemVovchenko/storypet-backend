package exceptions

import "errors"

var (
	UnprocessableURIParam = errors.New("unprocessable uri parameter was providedF")
	RecordAlreadyExist    = errors.New("can't create, record already exist")

	DatabaseDumpFailed        = errors.New("could not create database dump")
	DumpFileNotFoundInRequest = errors.New("no file field found in form/multipart section of request")
	DumpSaveFailed            = errors.New("can not save specified file")

	IncorrectAuthData     = errors.New("incorrect email or password")
	IncorrectRefreshToken = errors.New("incorrect refresh token")

	RequestedUserNotFound = errors.New("no user with requested id")
	IncorrectOldPassword  = errors.New("incorrect old password")
	UserIsNotVeterinarian = errors.New("operation permitted, selected user is not veterinarian")

	CanNotUpdatePrimaryRole = errors.New("could not update base system roles: administrator, subscribed/unsubscribed user and veterinarian")
	CanNotDeletePrimaryRole = errors.New("could not delete base system roles: administrator, subscribed/unsubscribed user and veterinarian")

	CanNotAssignPetToAnotherUser = errors.New("assigning new pet to other user is permitted")
	PetHasVeterinarian           = errors.New("can not assign veterinarian to pet, pet currently has a veterinarian")
	PetHasNoVeterinarian         = errors.New("pet has not a veterinarian")
	NoParentsSpecified           = errors.New("no parents specified")
)
