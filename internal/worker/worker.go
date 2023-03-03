package worker

import (
	"errors"
	"sync"
	"time"

	"github.com/22Fariz22/loyal/internal/config"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/pkg/logger"
)

type Pool struct {
	l          logger.Interface
	wg         sync.WaitGroup
	once       sync.Once
	shutDown   chan struct{}
	mainCh     chan workerData
	repository UseCase
	cfg        *config.Config
}

func NewWorkerPool(repo UseCase, l logger.Interface, cfg *config.Config) *Pool {
	return &Pool{
		l:          l,
		wg:         sync.WaitGroup{},
		once:       sync.Once{},
		shutDown:   make(chan struct{}),
		mainCh:     make(chan workerData, 10),
		repository: repo,
		cfg:        cfg,
	}
}

type workerData struct {
	orders []*entity.Order
}

//  функция которая каждые 2 мин забирает из таблицы ордеры со статусом NEW и кладет их в каналы
func CollectNewOrders(uc UseCase, l logger.Interface, cfg *config.Config) []*entity.Order {
	workers := NewWorkerPool(uc, l, cfg)

	workers.RunWorkers(5)
	defer workers.Stop()

	for {
		time.Sleep(10 * time.Second)

		newOrders, err := workers.repository.CheckNewOrders(l) //получаем список новых ордеров
		if err != nil {
			l.Info("err in CheckNewOrders(): ", err)
		}

		workers.AddJob(newOrders)
	}
}

type NewOrders struct {
	number string
}

func (w *Pool) AddJob(arr []*entity.Order) error {
	select {
	case <-w.shutDown:
		return errors.New("all channels are closed")
	case w.mainCh <- workerData{
		orders: arr,
	}:
		return nil
	}
}

// далее этот список передаем воркеру
func (w *Pool) RunWorkers(count int) {
	for i := 0; i < count; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for {
				select {
				case <-w.shutDown:
					return
				case orders, ok := <-w.mainCh:
					if !ok {
						return
					}
					_, err := w.repository.SendToAccrualBox(w.l, w.cfg, orders.orders)
					if err != nil {
						w.l.Info("err in SendToAccrualBox():", err)
					}
				}
			}
		}()
	}
}

func (w *Pool) Stop() {
	w.once.Do(func() {
		close(w.shutDown)
		close(w.mainCh)
	})
	w.wg.Wait()
}
