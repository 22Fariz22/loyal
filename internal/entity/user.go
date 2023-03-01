package entity

type User struct {
	ID            string
	Login         string
	Password      string
	BalanceTotal  int
	WithdrawTotal int
}
