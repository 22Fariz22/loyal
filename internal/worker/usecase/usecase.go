package usecase

import (
	"github.com/22Fariz22/loyal/internal/config"
	"github.com/22Fariz22/loyal/internal/entity"
	"github.com/22Fariz22/loyal/internal/worker"
	"github.com/22Fariz22/loyal/pkg/logger"
	"log"
)

type WorkerUseCase struct {
	WorkerRepo worker.WorkerRepository
}

func NewWorkerUseCase(repo worker.WorkerRepository) *WorkerUseCase {
	return &WorkerUseCase{
		WorkerRepo: repo,
	}
}

func (w *WorkerUseCase) CheckNewOrders(l logger.Interface) ([]*entity.Order, error) {
	log.Println("worker-uc-CheckNewOrders().")
	//return w.CheckNewOrders(l)
	return w.WorkerRepo.CheckNewOrders(l)

}

func (w *WorkerUseCase) SendToAccrualBox(l logger.Interface, cfg *config.Config, orders []*entity.Order) ([]*entity.History, error) {
	log.Println("worker-uc-SendToAccrualBox().")
	return w.WorkerRepo.SendToAccrualBox(l, cfg, orders)
}

//func (w *WorkerUseCase) SendToWaitListChannels() {
//	//TODO implement me
//	panic("implement me")
//}
