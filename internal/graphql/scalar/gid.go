// internal/graphql/scalar/gid.go
package scalar

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/dfodeker/storeos/internal/domain"
)

// MarshalGID converts a domain.GID to GraphQL ID
func MarshalGID(gid domain.GID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, strconv.Quote(gid.Encode()))
	})
}

// UnmarshalGID converts a GraphQL ID to domain.GID
func UnmarshalGID(v interface{}) (domain.GID, error) {
	switch v := v.(type) {
	case string:
		return domain.ParseGID(v)
	default:
		return domain.GID{}, fmt.Errorf("GID must be a string")
	}
}
