// internal/graphql/scalar/money.go
package scalar

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

// Money is cents stored as int64
type Money int64

func MarshalMoney(m Money) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.FormatInt(int64(m), 10))
	})
}

func UnmarshalMoney(v interface{}) (Money, error) {
	switch v := v.(type) {
	case int:
		return Money(v), nil
	case int64:
		return Money(v), nil
	case float64:
		return Money(int64(v)), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		return Money(i), err
	default:
		return 0, fmt.Errorf("Money must be an integer")
	}
}
