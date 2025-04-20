package money

import (
	"errors"
	"math"
)

// Injection points for backward compatibility.
// If you need to keep your JSON marshal/unmarshal way, overwrite them like below.
//
//	money.UnmarshalJSON = func (m *Money, b []byte) error { ... }
//	money.MarshalJSON = func (m Money) ([]byte, error) { ... }
var (
	// ErrCurrencyMismatch happens when two compared Money don't have the same Currency.
	ErrCurrencyMismatch = errors.New("currencies don't match")

	// ErrInvalidJSONUnmarshal happens when the default money.UnmarshalJSON fails to unmarshal Money because of invalid data.
	ErrInvalidJSONUnmarshal = errors.New("invalid json unmarshal")
)

// Amount is a data structure that stores the Amount being used for calculations.
type Amount = int64

// Money represents monetary value information, stores
// Currency and Amount value.
type Money struct {
	Amount   Amount    `json:"amount" bson:"amount"`
	Currency *Currency `json:"currency" bson:"currency"`
}

// New creates and returns new instance of Money.
func New(amount int64, code string) *Money {
	return &Money{
		Amount:   amount,
		Currency: newCurrency(code).get(),
	}
}

// NewFromFloat creates and returns new instance of Money from a float64.
// Always rounding trailing decimals down.
func NewFromFloat(amount float64, currency string) *Money {
	currencyDecimals := math.Pow10(GetCurrency(currency).Fraction)
	return New(int64(math.Round(amount*currencyDecimals)), currency)
}

// SameCurrency check if given Money is equals by Currency.
func (m *Money) SameCurrency(om *Money) bool {
	return m.Currency.equals(om.Currency)
}

func (m *Money) assertSameCurrency(om *Money) error {
	if !m.SameCurrency(om) {
		return ErrCurrencyMismatch
	}

	return nil
}

func (m *Money) compare(om *Money) int {
	switch {
	case m.Amount > om.Amount:
		return 1
	case m.Amount < om.Amount:
		return -1
	}

	return 0
}

// Equals checks equality between two Money types.
func (m *Money) Equals(om *Money) (bool, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return false, err
	}

	return m.compare(om) == 0, nil
}

// GreaterThan checks whether the value of Money is greater than the other.
func (m *Money) GreaterThan(om *Money) (bool, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return false, err
	}

	return m.compare(om) == 1, nil
}

// GreaterThanOrEqual checks whether the value of Money is greater or equal than the other.
func (m *Money) GreaterThanOrEqual(om *Money) (bool, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return false, err
	}

	return m.compare(om) >= 0, nil
}

// LessThan checks whether the value of Money is less than the other.
func (m *Money) LessThan(om *Money) (bool, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return false, err
	}

	return m.compare(om) == -1, nil
}

// LessThanOrEqual checks whether the value of Money is less or equal than the other.
func (m *Money) LessThanOrEqual(om *Money) (bool, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return false, err
	}

	return m.compare(om) <= 0, nil
}

// IsZero returns boolean of whether the value of Money is equals to zero.
func (m *Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive returns boolean of whether the value of Money is positive.
func (m *Money) IsPositive() bool {
	return m.Amount > 0
}

// IsNegative returns boolean of whether the value of Money is negative.
func (m *Money) IsNegative() bool {
	return m.Amount < 0
}

// Absolute returns new Money struct from given Money using absolute monetary value.
func (m *Money) Absolute() *Money {
	return &Money{Amount: mutate.calc.absolute(m.Amount), Currency: m.Currency}
}

// Negative returns new Money struct from given Money using negative monetary value.
func (m *Money) Negative() *Money {
	return &Money{Amount: mutate.calc.negative(m.Amount), Currency: m.Currency}
}

// Add returns new Money struct with value representing sum of Self and Other Money.
func (m *Money) Add(om *Money) (*Money, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return nil, err
	}

	return &Money{Amount: mutate.calc.add(m.Amount, om.Amount), Currency: m.Currency}, nil
}

func (m *Money) AddByInt64(val int64) *Money {
	mv := New(val, m.Currency.Code)
	return &Money{Amount: mutate.calc.add(m.Amount, mv.Amount), Currency: m.Currency}
}

func (m *Money) AddByFloat64(val float64) *Money {
	mv := NewFromFloat(val, m.Currency.Code)
	return &Money{Amount: mutate.calc.add(m.Amount, mv.Amount), Currency: m.Currency}
}

// Subtract returns new Money struct with value representing difference of Self and Other Money.
func (m *Money) Subtract(om *Money) (*Money, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return nil, err
	}

	return &Money{Amount: mutate.calc.subtract(m.Amount, om.Amount), Currency: m.Currency}, nil
}

func (m *Money) SubtractByInt64(val int64) *Money {
	mv := New(val, m.Currency.Code)
	return &Money{Amount: mutate.calc.subtract(m.Amount, mv.Amount), Currency: m.Currency}
}

func (m *Money) SubtractByFloat64(val float64) *Money {
	mv := NewFromFloat(val, m.Currency.Code)
	return &Money{Amount: mutate.calc.subtract(m.Amount, mv.Amount), Currency: m.Currency}
}

// Multiply returns new Money struct with value representing Self multiplied value by multiplier.
func (m *Money) Multiply(mul int64) *Money {
	return &Money{Amount: mutate.calc.multiply(m.Amount, mul), Currency: m.Currency}
}

func (m *Money) Divide(div int64) *Money {
	return &Money{Amount: mutate.calc.divide(m.Amount, div), Currency: m.Currency}
}

// Round returns new Money struct with value rounded to nearest zero.
func (m *Money) Round() *Money {
	return &Money{Amount: mutate.calc.round(m.Amount, m.Currency.Fraction), Currency: m.Currency}
}

// Split returns slice of Money structs with split Self value in given number.
// After division leftover pennies will be distributed round-robin amongst the parties.
// This means that parties listed first will likely receive more pennies than ones that are listed later.
func (m *Money) Split(n int) ([]*Money, error) {
	if n <= 0 {
		return nil, errors.New("split must be higher than zero")
	}

	a := mutate.calc.divide(m.Amount, int64(n))
	ms := make([]*Money, n)

	for i := 0; i < n; i++ {
		ms[i] = &Money{Amount: a, Currency: m.Currency}
	}

	r := mutate.calc.modulus(m.Amount, int64(n))
	l := mutate.calc.absolute(r)
	// Add leftovers to the first parties.

	v := int64(1)
	if m.Amount < 0 {
		v = -1
	}
	for p := 0; l != 0; p++ {
		ms[p].Amount = mutate.calc.add(ms[p].Amount, v)
		l--
	}

	return ms, nil
}

// Allocate returns slice of Money structs with split Self value in given ratios.
// It lets split money by given ratios without losing pennies and as Split operations distributes
// leftover pennies amongst the parties with round-robin principle.
func (m *Money) Allocate(rs ...int) ([]*Money, error) {
	if len(rs) == 0 {
		return nil, errors.New("no ratios specified")
	}

	// Calculate sum of ratios.
	var sum uint
	for _, r := range rs {
		if r < 0 {
			return nil, errors.New("negative ratios not allowed")
		}
		sum += uint(r)
	}

	var total int64
	ms := make([]*Money, 0, len(rs))
	for _, r := range rs {
		party := &Money{
			Amount:   mutate.calc.allocate(m.Amount, uint(r), sum),
			Currency: m.Currency,
		}

		ms = append(ms, party)
		total += party.Amount
	}

	// if the sum of all ratios is zero, then we just returns zeros and don't do anything
	// with the leftover
	if sum == 0 {
		return ms, nil
	}

	// Calculate leftover value and divide to first parties.
	lo := m.Amount - total
	sub := int64(1)
	if lo < 0 {
		sub = -sub
	}

	for p := 0; lo != 0; p++ {
		ms[p].Amount = mutate.calc.add(ms[p].Amount, sub)
		lo -= sub
	}

	return ms, nil
}

// Display lets represent Money struct as string in given Currency value.
func (m *Money) Display() string {
	c := m.Currency.get()
	return c.Formatter().Format(m.Amount)
}

// AsMajorUnits lets represent Money struct as subunits (float64) in given Currency value
func (m *Money) AsMajorUnits() float64 {
	c := m.Currency.get()
	return c.Formatter().ToMajorUnits(m.Amount)
}

// Compare function compares two money of the same type
//
//	if m.Amount > om.Amount returns (1, nil)
//	if m.Amount == om.Amount returns (0, nil
//	if m.Amount < om.Amount returns (-1, nil)
//
// If compare moneys from distinct Currency, return (m.Amount, ErrCurrencyMismatch)
func (m *Money) Compare(om *Money) (int, error) {
	if err := m.assertSameCurrency(om); err != nil {
		return int(m.Amount), err
	}

	return m.compare(om), nil
}
