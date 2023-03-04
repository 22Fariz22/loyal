package usecase

import (
	"context"
	"time"

	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/internal/order"
	"github.com/22Fariz22/loyal/pkg/logger"
)

type OrderUseCase struct {
	orderRepo order.OrderRepository
}

func NewOrderUseCase(orderRepo order.OrderRepository) *OrderUseCase {
	return &OrderUseCase{orderRepo: orderRepo}
}

func (o *OrderUseCase) PushOrder(ctx context.Context, l logger.Interface, user *entity.User, number string) error {
	eo := &entity.Order{ // можно ли убрать это или перенести это действие в репо?
		//ID:         "",
		UserID:     user.ID,
		Number:     number,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}
	return o.orderRepo.PushOrder(ctx, l, user, eo)
}

func (o *OrderUseCase) GetOrders(ctx context.Context, l logger.Interface, user *entity.User) ([]*entity.Order, error) {
	orders, err := o.orderRepo.GetOrders(ctx, l, user)
	if err != nil {
		l.Error("order-uc-GetOrders() -err: ", err)
		return nil, err
	}
	return orders, nil
}
