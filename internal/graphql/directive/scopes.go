// internal/graphql/directive/scopes.go
package directive

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"

	"yourapp/internal/domain"
	"yourapp/internal/http/middleware"
)

type DirectiveConfig struct{}

func NewDirectiveConfig() DirectiveConfig {
	return DirectiveConfig{}
}

// RequireScope directive implementation
func (d DirectiveConfig) RequireScope(ctx context.Context, obj interface{}, next graphql.Resolver, scope string) (interface{}, error) {
	auth := middleware.AuthFromContext(ctx)
	if auth == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	requiredScope := domain.Scope(scope)
	if !auth.Scopes.Has(requiredScope) {
		return nil, fmt.Errorf("access denied: missing scope %s", scope)
	}

	return next(ctx)
}

// RequireAnyScope directive implementation
func (d DirectiveConfig) RequireAnyScope(ctx context.Context, obj interface{}, next graphql.Resolver, scopes []string) (interface{}, error) {
	auth := middleware.AuthFromContext(ctx)
	if auth == nil {
		return nil, fmt.Errorf("unauthorized")
	}

	for _, s := range scopes {
		if auth.Scopes.Has(domain.Scope(s)) {
			return next(ctx)
		}
	}

	return nil, fmt.Errorf("access denied: requires one of %v", scopes)
}
