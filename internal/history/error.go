package history

import "errors"

var (
	ErrNotEnoughFunds      = errors.New("not enough funds")
	ErrInvalidOrderNumber  = errors.New("invalid order number")
	ErrThereIsNoWithdrawal = errors.New("there is no withdrawal")
)
