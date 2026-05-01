package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type ProfileMetadata struct {
	Namespace       string
	AppLabel        string
	CoDriverJobName string
	Filename        string
}

type Manager struct {
	basePath   string
	dateFormat string
}

func NewManager(basePath, dateFormat string) (*Manager, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	if dateFormat == "" {
		return nil, fmt.Errorf("dateFormat is required")
	}

	return &Manager{
		basePath:   basePath,
		dateFormat: dateFormat,
	}, nil
}

func (m *Manager) SaveProfile(r io.Reader, metadata ProfileMetadata) error {
	datePath := time.Now().Format(m.dateFormat)
	path := filepath.Join(
		m.basePath,
		metadata.Namespace,
		metadata.AppLabel,
		metadata.CoDriverJobName,
		datePath,
		metadata.Filename,
	)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			_ = err
		}
	}()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write profile data: %w", err)
	}

	return nil
}
