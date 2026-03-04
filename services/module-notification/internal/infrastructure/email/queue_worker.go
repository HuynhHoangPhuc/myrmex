package email

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/HuynhHoangPhuc/myrmex/services/module-notification/internal/infrastructure/persistence"
)

const (
	workerBatchSize    = 20
	workerPollInterval = 15 * time.Second
)

// QueueWorker polls the email_queue table and dispatches pending emails via SMTP.
type QueueWorker struct {
	repo *persistence.EmailQueueRepository
	smtp *SMTPService
	log  *zap.Logger
}

// NewQueueWorker creates a QueueWorker. If smtp is nil, it becomes a no-op.
func NewQueueWorker(repo *persistence.EmailQueueRepository, smtp *SMTPService, log *zap.Logger) *QueueWorker {
	return &QueueWorker{repo: repo, smtp: smtp, log: log}
}

// Start runs the polling loop until ctx is cancelled.
func (w *QueueWorker) Start(ctx context.Context) {
	if w.smtp == nil {
		w.log.Info("email queue worker: SMTP not configured, worker disabled")
		return
	}
	w.log.Info("email queue worker started", zap.Duration("interval", workerPollInterval))
	ticker := time.NewTicker(workerPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("email queue worker stopped")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *QueueWorker) processBatch(ctx context.Context) {
	rows, err := w.repo.FetchPending(ctx, workerBatchSize)
	if err != nil {
		w.log.Error("fetch pending emails failed", zap.Error(err))
		return
	}
	for _, row := range rows {
		w.processOne(ctx, row)
	}
}

func (w *QueueWorker) processOne(ctx context.Context, row persistence.EmailQueueRow) {
	err := w.smtp.Send(row.RecipientEmail, row.Subject, row.HTMLBody)
	if err == nil {
		if markErr := w.repo.MarkSent(ctx, row.ID); markErr != nil {
			w.log.Error("mark sent failed", zap.String("id", row.ID), zap.Error(markErr))
		}
		w.log.Debug("email sent", zap.String("id", row.ID), zap.String("to", row.RecipientEmail))
		return
	}

	w.log.Warn("email send failed", zap.String("id", row.ID), zap.Error(err))

	if row.RetryCount+1 >= row.MaxRetries {
		if failErr := w.repo.MarkFailed(ctx, row.ID, err.Error()); failErr != nil {
			w.log.Error("mark failed error", zap.String("id", row.ID), zap.Error(failErr))
		}
		w.log.Error("email permanently failed", zap.String("id", row.ID), zap.String("to", row.RecipientEmail))
		return
	}

	if retryErr := w.repo.IncrementRetry(ctx, row.ID, err.Error(), row.RetryCount); retryErr != nil {
		w.log.Error("increment retry failed", zap.String("id", row.ID), zap.Error(retryErr))
	}
}
