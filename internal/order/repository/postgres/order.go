package postgres

import (
	"context"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/internal/order"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/22Fariz22/loyal/pkg/postgres"
	"log"
	"strconv"
	"time"
)

type OrderRepository struct {
	*postgres.Postgres
}

type Order struct {
	ID         string
	UserID     string
	Number     string
	Status     string
	Accrual    uint32
	UploadedAt time.Time
}

func NewOrderRepository(db *postgres.Postgres) *OrderRepository {
	return &OrderRepository{db}
}

//type existOrder struct {
//	uID    int    `json:"user_id"`
//	number string `json:"number"`
//}

func (o *OrderRepository) PushOrder(ctx context.Context, l logger.Interface, user *entity.User, eo *entity.Order) error {
	log.Println("order-repo-PushOrder().")

	var existUser int

	log.Println("order-repo-PushOrder()-number:", eo.Number)

	_ = o.Pool.QueryRow(ctx, `SELECT user_id FROM orders where number = $1;`, eo.Number).Scan(&existUser)

	existUserConvToStr, err := strconv.Atoi(user.ID)
	if err != nil {
		l.Error("err in strconv.Atoi(user.ID):", err)
		return err
	}

	log.Println(" existUser == existUserConvToStr", existUser, existUserConvToStr)
	if existUser == existUserConvToStr {
		l.Info("Number Has Already Been Uploaded")
		return order.ErrNumberHasAlreadyBeenUploaded
	}

	if existUser != existUserConvToStr {
		var numbExist string
		_ = o.Pool.QueryRow(ctx, `SELECT number FROM orders where number = $1;`, eo.Number).Scan(&numbExist)

		if numbExist == eo.Number {
			l.Info("Number Has Already Been Uploaded By AnotherUser")
			return order.ErrNumberHasAlreadyBeenUploadedByAnotherUser
		}
	}

	log.Println("order-repo-PushOrder()- start begin tx.")
	tx, err := o.Pool.Begin(ctx)
	if err != nil {
		l.Error("tx err: ", err)
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO orders (user_id, number, order_status, uploaded_at)
								VALUES ($1,$2,$3,$4)`,
		eo.UserID, eo.Number, eo.Status, eo.UploadedAt)
	if err != nil {
		l.Error("order-repo-PushOrder()-err(1): ", err)
		return err
	}

	_, err = tx.Exec(ctx, `INSERT INTO history (user_id, number,processed_at)
								VALUES ($1,$2,$3)`,
		eo.UserID, eo.Number, eo.UploadedAt)
	if err != nil {
		l.Error("order-repo-PushOrder()-err(2): ", err)
		return err
	}

	log.Println("order-repo-PushOrder().-tx commit.")
	err = tx.Commit(ctx)
	if err != nil {
		l.Error("commit err: ", err)
		return err
	}

	return nil
}

func (o OrderRepository) GetOrders(ctx context.Context, l logger.Interface, user *entity.User) ([]*entity.Order, error) {
	log.Println("order-repo-GetOrder().")

	rows, err := o.Pool.Query(ctx, `SELECT order_id, number, order_status, accrual, uploaded_at FROM orders
									WHERE user_id = $1`, user.ID)
	if err != nil {
		l.Error("order-repo-GetOrders()-Pool.Query()-err: ", err)
		return nil, err
	}

	out := make([]*entity.Order, 0)

	for rows.Next() {
		order := new(entity.Order)

		err := rows.Scan(&order.ID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			l.Error("order-repo-GetOrders()-rows.Scan()-err: ", err)
			return nil, err
		}
		out = append(out, order)
	}

	if len(out) == 0 {
		l.Info("order-repo-GetOrders()-len(out)==0")
		return nil, order.ErrThereIsNoOrders
	}

	return out, nil
}

func toModel(o *entity.Order) *Order {
	return &Order{
		ID:         o.ID,
		UserID:     o.UserID,
		Number:     o.Number,
		Status:     o.Status,
		UploadedAt: o.UploadedAt,
	}
}
