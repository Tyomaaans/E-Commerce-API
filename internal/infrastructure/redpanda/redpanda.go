package redpanda

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Config untuk Redpanda broker.
// Di-inject dari env: REDPANDA_BROKERS=localhost:9092
type Config struct {
	Brokers []string // e.g. ["localhost:9092"]
}

func ParseBrokers(raw string) []string {
	brokers := strings.Split(raw, ",")
	result := make([]string, 0, len(brokers))
	for _, b := range brokers {
		b = strings.TrimSpace(b)
		if b != "" {
			result = append(result, b)
		}
	}
	return result
}

// EnsureTopics memastikan topic yang dibutuhkan sudah ada.
// Dipanggil saat startup sebelum producer/consumer diinisialisasi.
func EnsureTopics(brokers []string, topics ...string) error {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return fmt.Errorf("redpanda: failed to connect: %w", err)
	}
	defer client.Close()

	adm := kadm.NewClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	existing, err := adm.ListTopics(ctx)
	if err != nil {
		return fmt.Errorf("redpanda: failed to list topics: %w", err)
	}

	var toCreate []string
	for _, topic := range topics {
		if _, ok := existing[topic]; !ok {
			toCreate = append(toCreate, topic)
		}
	}

	if len(toCreate) == 0 {
		return nil
	}

	responses, err := adm.CreateTopics(ctx, 3, 1, nil, toCreate...)
	if err != nil {
		return fmt.Errorf("redpanda: failed to create topics: %w", err)
	}

	for _, r := range responses {
		if r.Err != nil {
			return fmt.Errorf("redpanda: failed to create topic %s: %w", r.Topic, r.Err)
		}
	}

	return nil
}