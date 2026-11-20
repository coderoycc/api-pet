package application

import (
	"context"

	"api-go/internal/audit/domain"
)

type actorContextKey struct{}

// WithActor adjunta la información del actor autenticado al context.Context.
func WithActor(ctx context.Context, actor domain.Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

// ActorFromContext extrae el actor de auditoría desde el context.Context si existe.
func ActorFromContext(ctx context.Context) (domain.Actor, bool) {
	if ctx == nil {
		return domain.Actor{}, false
	}
	actor, ok := ctx.Value(actorContextKey{}).(domain.Actor)
	return actor, ok
}
