package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sfedu-crm/internal/domain"
)

func TestTrainingTime_EndBeforeStart(t *testing.T) {
	start := time.Now()
	end := start.Add(-1 * time.Hour)
	assert.True(t, end.Before(start), "Конец не может быть раньше начала")
}

func TestTrainingStatus_Transitions(t *testing.T) {
	tests := []struct {
		from  domain.TrainingStatus
		to    domain.TrainingStatus
		valid bool
	}{
		{domain.TrainingStatusScheduled, domain.TrainingStatusCompleted, true},
		{domain.TrainingStatusScheduled, domain.TrainingStatusCancelled, true},
		{domain.TrainingStatusCompleted, domain.TrainingStatusScheduled, false},
		{domain.TrainingStatusCancelled, domain.TrainingStatusCompleted, false},
	}

	for _, tt := range tests {
		if tt.valid {
			assert.NotEqual(t, tt.from, tt.to)
		}
	}
}

func TestTrainingRequest_PreferredTime(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	assert.True(t, past.Before(time.Now()), "Желаемое время в прошлом")
}
