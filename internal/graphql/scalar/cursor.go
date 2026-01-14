package scalar

// internal/graphql/scalar/cursor.go

import (
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

type Cursor string

func EncodeCursor(value string) Cursor {
	return Cursor(base64.StdEncoding.EncodeToString([]byte(value)))
}

func (c Cursor) Decode() (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(c))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func MarshalCursor(c Cursor) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(string(c)))
	})
}

func UnmarshalCursor(v interface{}) (Cursor, error) {
	switch v := v.(type) {
	case string:
		return Cursor(v), nil
	default:
		return "", fmt.Errorf("Cursor must be a string")
	}
}
