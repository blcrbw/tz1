package subscription

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, subscription *Subscription) error
	FindAll(ctx context.Context) (s []Subscription, err error)
	GetList(ctx context.Context, limit int, offset int, form string, to string, user string, service string) (s []Subscription, err error)
	GetSum(ctx context.Context, form string, to string, user string, service string) (sum int64, err error)
	FindOne(ctx context.Context, id string) (Subscription, error)
	Update(ctx context.Context, id string, subscription *Subscription) error
	Delete(ctx context.Context, id string) error
}
