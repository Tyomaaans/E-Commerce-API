package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// ==== Email Job Types ====

type EmailJobType string

const (
	EmailJobVerifyRegister        EmailJobType = "verify_register"
	EmailJobRegisterSuccess       EmailJobType = "register_success"
	EmailJobResetPassword         EmailJobType = "reset_password"
	EmailJobUpdatePasswordSuccess EmailJobType = "update_password_success"
	EmailJobRegisterStoreSuccess  EmailJobType = "register_store_success"
)

const emailTopic = "email-jobs"

// EmailJob adalah struktur pesan yang dikirim ke Redpanda topic.
type EmailJob struct {
	Type  EmailJobType `json:"type"`
	To    string       `json:"to"`
	Name  string       `json:"name,omitempty"`  // Digunakan untuk salam/sapaan
	Token string       `json:"token,omitempty"` // Hanya dipakai jika job butuh link/token
}
// ==== Producer Interface ====

type EmailProducer interface {
	PublishEmailJob(ctx context.Context, job EmailJob) error
	Close()
}

// ==== Implementation ====

type redpandaEmailProducer struct {
	client *kgo.Client
}

func NewRedpandaEmailProducer(brokers []string) (EmailProducer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		// Redpanda-compatible: tidak perlu konfigurasi khusus
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create redpanda producer: %w", err)
	}

	return &redpandaEmailProducer{client: client}, nil
}

func (p *redpandaEmailProducer) PublishEmailJob(ctx context.Context, job EmailJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal email job: %w", err)
	}

	record := &kgo.Record{
		Topic: emailTopic,
		Key:   []byte(job.To), // partisi by email untuk ordering per user
		Value: payload,
	}

	// SyncProduce: tunggu konfirmasi dari broker
	results := p.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("failed to publish email job: %w", err)
	}

	return nil
}

func (p *redpandaEmailProducer) Close() {
	p.client.Close()
}