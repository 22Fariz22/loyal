package postgres

import (
	"context"
	"fmt"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/internal/history"
	"github.com/22Fariz22/loyal/pkg/logger"
	"github.com/22Fariz22/loyal/pkg/postgres"
	"github.com/georgysavva/scany/v2/pgxscan"
	"log"
	"reflect"
)

type HistoryRepository struct {
	*postgres.Postgres
}

func NewHistoryRepository(db *postgres.Postgres) *HistoryRepository {
	return &HistoryRepository{db}
}

type UserBalance struct {
	BalanceTotal  int
	WithdrawTotal int
}

func (h *HistoryRepository) GetBalance(ctx context.Context, l logger.Interface, user *entity.User) (*entity.User, error) {
	fmt.Println("history-repo-GetBalance()-user: ", user)
	var ub UserBalance

	var u entity.User

	err := pgxscan.Get(ctx, h.Pool, &ub, `SELECT balance_total, withdraw_total FROM users where user_id = $1;`, user.ID)
	if err != nil {
		l.Error("history-repo-GetBalance()-err: ", err)
		return nil, err
	}
	fmt.Println("history-repo-GetBalance()-ub: ", ub)

	u.BalanceTotal = ub.BalanceTotal
	u.WithdrawTotal = ub.WithdrawTotal

	return &u, nil
}

func (h *HistoryRepository) Withdraw(ctx context.Context, l logger.Interface, user *entity.User,
	number string, withdrawResp int) error {
	log.Println("hist-repo-Withdraw().")
	withdrawTotal := 0

	// узнаем сколько всего баллов
	err := pgxscan.Get(ctx, h.Pool, &withdrawTotal, `SELECT balance_total FROM users WHERE user_id = $1`, user.ID)

	if err != nil {
		l.Error("history-repo-Get()-err: ", err)
		return err
	}
	log.Println("hist-repo-Withdraw()-withdrawTotal: ", withdrawTotal, "refl: ", reflect.TypeOf(withdrawTotal))
	log.Println("hist-repo-Withdraw()-withdrawResp: ", withdrawResp, "refl: ", reflect.TypeOf(withdrawResp))

	//сравниваем наш баланс с запросом
	if withdrawTotal < withdrawResp || withdrawResp < 0 {
		l.Error("history-repo-Withdraw()- withdraw_total<withdrawResp): ", history.ErrNotEnoughFunds)
		return history.ErrNotEnoughFunds
	}

	log.Println("hist-repo-Withdraw()-start tx begin.")
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		l.Error("tx err: ", err)
		return err
	}
	defer tx.Rollback(ctx)

	//UPDATE в таблице USER полей balance_total и withdraw_total
	_, err = tx.Prepare(ctx, "UPDATE", `UPDATE users SET balance_total = balance_total - $1,
								withdraw_total = withdraw_total + $1 WHERE user_id = $2;`)
	if err != nil {
		l.Error("error in tx.Prepare UPDATE: ", err)
	}
	_, err = tx.Exec(ctx, `UPDATE users SET balance_total = balance_total - $1,
						   withdraw_total = withdraw_total + $1 WHERE user_id = $2;`, withdrawResp, user.ID)
	if err != nil {
		l.Error("error in tx.Exec UPDATE: ", err)
		return err
	}

	// INSERT в таблицу history
	_, err = tx.Prepare(ctx, "INSERT", `INSERT INTO history(user_id, number, sum) VALUES($1, $2, $3);`)
	if err != nil {
		l.Error("tx.Prepare INSERT: ", err)
		return err
	}

	log.Println("hist-repo-Withdraw()-before INSERT INTO history values user.ID, number, withdrawResp:", user.ID, number, withdrawResp)
	_, err = tx.Exec(ctx, `INSERT INTO history(user_id, number, sum) VALUES($1, $2, $3)`, user.ID, number, withdrawResp)
	if err != nil {
		l.Error("tx.Exec INSERT: ", err)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		l.Error("commit err: ", err)
		return err
	}
	log.Println("hist-repo-Withdraw()-end tx.Commit().")

	return nil
}

type histWithdrawalResp struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func (h *HistoryRepository) InfoWithdrawal(ctx context.Context, l logger.Interface,
	user *entity.User) ([]*entity.History, error) {
	log.Println("hist-repo-InfoWithdrawal().")

	rows, err := h.Pool.Query(ctx, `SELECT number, sum, processed_at FROM history WHERE user_id = $1 ORDER BY processed_at desc`,
		user.ID)
	if err != nil {
		l.Error("hist-repo-InfoWithdrawal()- err in Query SELECT: ", err)
		return nil, err
	}

	out := make([]*entity.History, 0)

	for rows.Next() {
		hist := new(entity.History)

		err := rows.Scan(&hist.Number, &hist.Sum, &hist.ProcessedAt)
		if err != nil {
			l.Error("error in rows.Scan(): ", err)
			return nil, err
		}
		log.Println("hist-repo-InfoWithdrawal()-rows.Next()-hist: ", hist)

		out = append(out, hist)
	}

	if len(out) == 0 {
		log.Println("hist-repo-InfoWithdrawal()-rows.Next()-len(out) == 0")
		return nil, history.ErrThereIsNoWithdrawal
	}
	log.Println("hist-repo-InfoWithdrawal()-out: ", out)

	return out, nil
}
