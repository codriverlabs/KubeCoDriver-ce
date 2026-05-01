package server

import (
	"context"
	"io"

	"github.com/codriverlabs/KubeCoDriver/pkg/collector/storage"

	authv1 "k8s.io/api/authentication/v1"
)

// StorageManager defines the interface for profile storage operations
type StorageManager interface {
	SaveProfile(r io.Reader, metadata storage.ProfileMetadata) error
}

// TokenValidator defines the interface for token validation operations
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*authv1.UserInfo, error)
}
