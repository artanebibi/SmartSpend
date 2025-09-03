package enum

type Currency string
type TransactionType string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	MKD Currency = "MKD"
)

const (
	Expense TransactionType = "Expense"
	Income  TransactionType = "Income"
)
