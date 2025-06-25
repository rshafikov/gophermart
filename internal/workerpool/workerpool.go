package workerpool

import (
	"context"
	"errors"
	"github.com/rshafikov/gophermart/internal/client"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/models"
	"github.com/rshafikov/gophermart/internal/schemas"
	"go.uber.org/zap"
	"time"
)

const MaxTasksInPool = 100
const TaskTimeout = time.Second

type balanceIncreaser interface {
	IncreaseUserBalance(ctx context.Context, userID int, amount float64) error
}

type orderUpdater interface {
	UpdateOrder(ctx context.Context, order *models.Order) error
}

type accrualClient interface {
	GetOrderStatus(ctx context.Context, number string) (*schemas.AccrualOrder, error)
}

type WorkerPool struct {
	workers        int
	tasksChan      chan OrderTask
	resultsChan    chan OrderTask
	stopChan       chan struct{}
	client         accrualClient
	orderService   orderUpdater
	balanceService balanceIncreaser
}

func NewWorkerPool(workers int, client accrualClient, ordersService orderUpdater, balanceService balanceIncreaser) *WorkerPool {
	return &WorkerPool{
		workers:        workers,
		tasksChan:      make(chan OrderTask, MaxTasksInPool),
		resultsChan:    make(chan OrderTask),
		stopChan:       make(chan struct{}),
		client:         client,
		orderService:   ordersService,
		balanceService: balanceService,
	}
}

func (wp *WorkerPool) Start() {
	for w := 1; w <= wp.workers; w++ {
		go wp.runWorker(w)
	}
	go wp.processResults()
	logger.L.Debug("worker pool started")
}

func (wp *WorkerPool) Stop() {
	close(wp.tasksChan)
	close(wp.resultsChan)
	close(wp.stopChan)
	logger.L.Debug("worker pool stopped")
}

func (wp *WorkerPool) processResults() {
	for r := range wp.resultsChan {
		if r.Error != nil {
			logger.L.Error("worker failed a task",
				zap.Int("worker_id", r.WorkerID),
				zap.String("order_id", r.OrderID),
				zap.Error(r.Error),
			)
			if errors.Is(r.Error, client.ErrTooManyRequests) {
				logger.L.Debug("too many requests, adding task back to worker pool", zap.String("order_id", r.OrderID))
			}
			if errors.Is(r.Error, client.ErrNoContent) {
				logger.L.Debug("order not found, adding task back to worker pool", zap.String("order_id", r.OrderID))
				continue
			}
			wp.AddTask(r)
		} else {
			logger.L.Debug("worker successfully finished a task",
				zap.Int("worker_id", r.WorkerID),
				zap.String("order_id", r.OrderID),
			)
			if r.LastAccrualStatus == schemas.AccrualOrderStatusProcessing || r.LastAccrualStatus == schemas.AccrualOrderStatusRegistered {
				logger.L.Debug("order status is processing or registered, adding task back to worker pool", zap.String("order_id", r.OrderID))
				wp.AddTask(r)
			}
		}
	}
}

func (wp *WorkerPool) runWorker(w int) {
	logger.L.Debug("worker started", zap.Int("worker", w))
	for {
		select {

		case <-wp.stopChan:
			logger.L.Debug("worker recieved stop signal", zap.Int("worker_id", w))
			return

		case task, ok := <-wp.tasksChan:
			if !ok {
				logger.L.Debug("jobs channel closed, closing worker", zap.Int("worker_id", w))
				return
			}
			logger.L.Debug("worker received job", zap.Int("worker_id", w), zap.String("order_id", task.OrderID))
			task.WorkerID = w
			wp.resultsChan <- wp.proceedOrder(task)

		}
	}
}

func (wp *WorkerPool) AddTask(task OrderTask) {
	logger.L.Debug("adding task to worker pool", zap.String("id", task.OrderID))
	wp.tasksChan <- task
}

func (wp *WorkerPool) proceedOrder(task OrderTask) OrderTask {
	ctx, cancel := context.WithTimeout(context.Background(), TaskTimeout)
	defer cancel()

	order, err := wp.client.GetOrderStatus(ctx, task.OrderID)

	if err != nil {
		logger.L.Error("unable to get order status", zap.Error(err))
		task.Error = err
		return task
	}

	if task.LastAccrualStatus == order.Status || order.Status == schemas.AccrualOrderStatusRegistered {
		logger.L.Debug("order status hasn't changed or it has been just registred", zap.String("order_id", task.OrderID))
		return task
	}
	task.LastAccrualStatus = order.Status

	var internalOrderStatus models.InternalOrderStatus
	switch order.Status {
	case schemas.AccrualOrderStatusInvalid:
		internalOrderStatus = models.OrderStatusInvalid
	case schemas.AccrualOrderStatusProcessed:
		internalOrderStatus = models.OrderStatusProcessed
	case schemas.AccrualOrderStatusProcessing:
		internalOrderStatus = models.OrderStatusProcessing
	}

	err = wp.orderService.UpdateOrder(ctx, &models.Order{
		NumeralID: task.OrderID,
		Status:    internalOrderStatus,
		Accrual:   order.Accrual,
	})
	if err != nil {
		logger.L.Error("unable to update order status in database", zap.Error(err))
		task.Error = err
		return task
	}

	if order.Status == schemas.AccrualOrderStatusProcessed {
		err = wp.balanceService.IncreaseUserBalance(ctx, task.UserID, order.Accrual)
		if err != nil {
			logger.L.Error("unable to increase user balance", zap.Error(err))
			task.Error = err
		}
	}

	return task
}
