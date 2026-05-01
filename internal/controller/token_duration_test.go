package controller

import (
	"context"
	"testing"
	"time"
)

func TestGetTokenDuration(t *testing.T) {
	r := &CoDriverJobReconciler{}
	ctx := context.Background()

	tests := []struct {
		name               string
		collectionDuration time.Duration
		wantMinimum        time.Duration
	}{
		{
			name:               "below minimum - 30 seconds",
			collectionDuration: 30 * time.Second,
			wantMinimum:        10 * time.Minute,
		},
		{
			name:               "below minimum - 5 minutes",
			collectionDuration: 5 * time.Minute,
			wantMinimum:        10 * time.Minute,
		},
		{
			name:               "at minimum - 10 minutes",
			collectionDuration: 10 * time.Minute,
			wantMinimum:        10 * time.Minute,
		},
		{
			name:               "above minimum - 15 minutes",
			collectionDuration: 15 * time.Minute,
			wantMinimum:        16 * time.Minute, // 15 + 1 minute buffer
		},
		{
			name:               "above minimum - 30 minutes",
			collectionDuration: 30 * time.Minute,
			wantMinimum:        31 * time.Minute, // 30 + 1 minute buffer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.getTokenDuration(ctx, tt.collectionDuration)
			if got < tt.wantMinimum {
				t.Errorf("getTokenDuration() = %v, want >= %v", got, tt.wantMinimum)
			}
		})
	}
}
