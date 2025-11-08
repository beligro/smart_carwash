package service

import (
	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/logger"
	"context"
	"fmt"
	"sync"
	"time"
)

// WebhookQueue очередь для асинхронной обработки webhook'ов
type WebhookQueue struct {
	tasks    chan *models.WebhookRequest
	wg       sync.WaitGroup
	service  *service
	workers  int
	shutdown chan struct{}
}

// NewWebhookQueue создает новую очередь для webhook'ов
func NewWebhookQueue(s *service, workers int) *WebhookQueue {
	if workers <= 0 {
		workers = 3 // По умолчанию 3 воркера
	}

	q := &WebhookQueue{
		tasks:    make(chan *models.WebhookRequest, 100), // Буфер на 100 webhook'ов
		service:  s,
		workers:  workers,
		shutdown: make(chan struct{}),
	}

	// Запускаем воркеры
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	logger.Printf("WebhookQueue: запущено %d воркеров", workers)
	return q
}

// worker обрабатывает webhook'и из очереди
func (q *WebhookQueue) worker(id int) {
	defer q.wg.Done()
	logger.Printf("WebhookWorker #%d: started", id)

	for {
		select {
		case task := <-q.tasks:
			if task == nil {
				// Канал закрыт
				logger.Printf("WebhookWorker #%d: остановлен", id)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

			start := time.Now()
			err := q.service.processWebhook(ctx, task)
			duration := time.Since(start)

			if err != nil {
				logger.Printf("WebhookWorker #%d: ERROR processing PaymentId=%d: %v (took %v)",
					id, task.PaymentId, err, duration)
			} else {
				logger.Printf("WebhookWorker #%d: SUCCESS PaymentId=%d (took %v)",
					id, task.PaymentId, duration)
			}

			cancel()

		case <-q.shutdown:
			logger.Printf("WebhookWorker #%d: получил сигнал остановки", id)
			return
		}
	}
}

// Enqueue добавляет webhook в очередь для обработки
func (q *WebhookQueue) Enqueue(req *models.WebhookRequest) error {
	select {
	case q.tasks <- req:
		logger.Printf("WebhookQueue: добавлен webhook PaymentId=%d в очередь", req.PaymentId)
		return nil
	default:
		return fmt.Errorf("webhook queue is full")
	}
}

// Shutdown останавливает очередь и все воркеры
func (q *WebhookQueue) Shutdown() {
	logger.Printf("WebhookQueue: начата остановка очереди")
	close(q.shutdown)
	close(q.tasks)
	q.wg.Wait()
	logger.Printf("WebhookQueue: все воркеры остановлены")
}
