package usecase

import (
	"context"
	"fmt"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/internal/order"
	"github.com/22Fariz22/loyal/pkg/logger"
	"log"
	"time"
)

type OrderUseCase struct {
	orderRepo order.OrderRepository
}

func NewOrderUseCase(orderRepo order.OrderRepository) *OrderUseCase {
	return &OrderUseCase{orderRepo: orderRepo}
}

func (o *OrderUseCase) PushOrder(ctx context.Context, l logger.Interface, user *entity.User, number string) error {
	log.Println("order-uc-PushOrder().")
	eo := &entity.Order{ // можно ли убрать это или перенести это действие в репо?
		//ID:         "",
		UserID:     user.ID,
		Number:     number,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}
	fmt.Println("orders-uc-eo.UserID: ", eo.UserID)
	return o.orderRepo.PushOrder(ctx, l, user, eo)
}

func (o *OrderUseCase) GetOrders(ctx context.Context, l logger.Interface, user *entity.User) ([]*entity.Order, error) {
	log.Println("order-uc-GetOrder().")
	orders, err := o.orderRepo.GetOrders(ctx, l, user)
	if err != nil {
		log.Println("order-uc-GetOrders() -err: ", err)
		return nil, err
	}
	return orders, nil
}
