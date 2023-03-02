package postgres

import (
	"context"
	"encoding/json"
	"github.com/22Fariz22/loyal/internal/config"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/22Fariz22/loyal/pkg/postgres"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

type WorkerRepository struct {
	*postgres.Postgres
}

func NewWorkerRepository(db *postgres.Postgres) *WorkerRepository {
	return &WorkerRepository{db}
}

func (w *WorkerRepository) CheckNewOrders(l logger.Interface) ([]*entity.Order, error) {
	ctx := context.Background()
	rows, err := w.Pool.Query(ctx, `SELECT user_id,number,order_status FROM orders
									WHERE order_status IN( 'NEW','PROCESSING')`)
	if err != nil {
		l.Error("err in Pool.Query()", err)
		return nil, err
	}

	out := make([]*entity.Order, 0)

	for rows.Next() {
		order := new(entity.Order)
		err := rows.Scan(&order.UserID, &order.Number, &order.Status)
		if err != nil {
			l.Error("err rows.Scan(): ", err)
			return nil, err
		}

		out = append(out, order)

	}
	return out, nil
}

// структура json ответа от accrual sysytem
type ResAccrualSystem struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

//SendToAccrualBox отправляем запрос accrual system и возвращаем ответ от него
func (w *WorkerRepository) SendToAccrualBox(l logger.Interface, cfg *config.Config, orders []*entity.Order) ([]*entity.History, error) {
	var resAccrSys ResAccrualSystem

	// считываем из env переменную ACCRUAL_SYSTEM_ADDRESS
	accrualSystemAddress := cfg.AccrualSystemAddress

	reqURL, err := url.Parse(accrualSystemAddress)
	if err != nil {
		l.Error("incorrect ACCRUAL_SYSTEM_ADDRESS:", err)
		return nil, err
	}

	// проходимся по списку ордеров и обращаемся к accrual system
	for _, v := range orders {
		uID, err := strconv.Atoi(v.UserID)
		if err != nil {
			l.Error("worker-repo-SendToAccrualBox()-atoi: ", err)
			return nil, err
		}

		reqURL.Path = path.Join("/api/orders/", v.Number)

		r, err := http.Get(reqURL.String())
		if err != nil {
			l.Error("can't do request: ", err)
			return nil, err //выходим из цикла, если не получился запрос к accrual system
		}

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			l.Error("Can't read response body: ", err)
			continue //переходим к следущей итерации
		}

		if r.StatusCode == 204 {
			if err := update(w, l, ResAccrualSystem{
				Order:   v.Number,
				Status:  "INVALID",
				Accrual: 0,
			}, uID); err != nil {
				return nil, err // определить какой error
			}
		}

		if r.StatusCode == 200 {
			err = json.Unmarshal(body, &resAccrSys)
			if err != nil {
				l.Error("Unmarshal error: ", err)
			}

			update(w, l, resAccrSys, uID)
		}

		if r.StatusCode == 429 {
			sleep, err := time.ParseDuration(r.Header.Get("Retry-After"))
			if err != nil {
				l.Error("worker-repo-SendToAccrualBox()-status 429- time.ParseDuration()-err: ", err)
				time.Sleep(60 * time.Second)
			}
			time.Sleep(sleep)
		}

		if r.StatusCode == 500 {
			l.Error("worker-repo-SendToAccrualBox()-status500.")
			return nil, err
		}
	}

	return nil, nil
}

func update(w *WorkerRepository, l logger.Interface, resAcc ResAccrualSystem, uID int) error {
	ctx := context.Background()

	//UPDATE в таблице History и Orders
	log.Println("worker-repo-updateWithStatus()- start begin tx.")
	tx, err := w.Pool.Begin(ctx)
	if err != nil {
		l.Error("tx err: ", err)
		return err
	}
	defer tx.Rollback(ctx)

	// добавлякем в таблицу orders
	_, err = tx.Exec(ctx, `UPDATE orders SET order_status =  $1, accrual = $2
							where number = $3`, resAcc.Status, int(resAcc.Accrual*100), resAcc.Order)
	if err != nil {
		l.Error("error in Exec UPDATE: ", err)
		return err
	}

	// добовляем в таблицу user
	_, err = tx.Exec(ctx, `UPDATE users SET balance_total =  $1
							where user_id = $2`, int(resAcc.Accrual*100), uID)
	if err != nil {
		l.Error("error in Exec UPDATE: ", err)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		l.Error("worker-repo-updateWithStatus() -tx.commit err: ", err)
		return err
	}

	return nil
}