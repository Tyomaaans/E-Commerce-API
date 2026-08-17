package users

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

// ==== Email Worker ====

// EmailWorker mengkonsumsi pesan dari Redpanda topic "email-jobs"
// dan mendelegasikan pengiriman email ke EmailService.
type EmailWorker struct {
	client       *kgo.Client
	emailService EmailService
	logger       *slog.Logger
	appURL       string
}

func NewEmailWorker(brokers []string, groupID string, appURL string, emailService EmailService, logger *slog.Logger) (*EmailWorker, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(emailTopic),
		// Auto-commit setelah proses berhasil
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create redpanda consumer: %w", err)
	}

	return &EmailWorker{
		client:       client,
		emailService: emailService,
		logger:       logger,
		appURL:       appURL,
	}, nil
}

// Run memulai consumer loop. Blokir sampai ctx dibatalkan.
func (w *EmailWorker) Run(ctx context.Context) {
	w.logger.Info("email worker started", "topic", emailTopic)

	for {
		fetches := w.client.PollFetches(ctx)

		if ctx.Err() != nil {
			w.logger.Info("email worker stopped")
			return
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fe := range errs {
				w.logger.Error("fetch error", "topic", fe.Topic, "partition", fe.Partition, "err", fe.Err)
			}
			continue
		}

		fetches.EachRecord(func(r *kgo.Record) {
			if err := w.processRecord(ctx, r); err != nil {
				w.logger.Error("failed to process email job",
					"key", string(r.Key),
					"offset", r.Offset,
					"err", err,
				)
				// Tidak commit offset — record akan di-retry saat restart
				// Untuk dead-letter queue, tambahkan logic publish ke DLQ topic di sini
				return
			}

			// Commit offset hanya setelah berhasil diproses
			if err := w.client.CommitRecords(ctx, r); err != nil {
				w.logger.Error("failed to commit offset", "offset", r.Offset, "err", err)
			}
		})
	}
}

func (w *EmailWorker) processRecord(ctx context.Context, r *kgo.Record) error {
	var job EmailJob
	if err := json.Unmarshal(r.Value, &job); err != nil {
		w.logger.Warn("malformed email job, skipping", "raw", string(r.Value))
		return nil
	}

	w.logger.Info("processing email job", "type", job.Type, "to", job.To)

	switch job.Type {
	case EmailJobVerifyRegister:
		// Susun link verifikasi sebelum memanggil email service
		link := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", w.appURL, job.Token)
		return w.emailService.SendVerificationEmail(ctx, job.To, link)

	case EmailJobRegisterSuccess:
		return w.emailService.SendRegisterSuccessEmail(ctx, job.To, job.Name)

	case EmailJobResetPassword:
		// Susun link reset password sebelum memanggil email service
		link := fmt.Sprintf("%s/api/v1/auth/reset-password?token=%s", w.appURL, job.Token)
		return w.emailService.SendResetPasswordEmail(ctx, job.To, link)

	case EmailJobUpdatePasswordSuccess:
		return w.emailService.SendUpdatePasswordSuccessEmail(ctx, job.To, job.Name)

	default:
		w.logger.Warn("unknown email job type, skipping", "type", job.Type)
		return nil
	}
}

func (w *EmailWorker) Close() {
	w.client.Close()
}