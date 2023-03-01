package order

import "errors"

var (
	//status 409
	ErrNumberHasAlreadyBeenUploadedByAnotherUser = errors.New("number has already been uploaded by another user")
	ErrNumberHasAlreadyBeenUploaded              = errors.New("number has already been uploaded")
	ErrThereIsNoOrders                           = errors.New("there is no orders")
)
