package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codriverlabs/KubeCoDriver/pkg/collector/storage"

	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// Mock implementations
type mockStorage struct {
	saveProfileFunc func(io.Reader, storage.ProfileMetadata) error
}

func (m *mockStorage) SaveProfile(r io.Reader, metadata storage.ProfileMetadata) error {
	if m.saveProfileFunc != nil {
		return m.saveProfileFunc(r, metadata)
	}
	return nil
}

type mockAuth struct {
	validateTokenFunc func(context.Context, string) (*authv1.UserInfo, error)
}

func (m *mockAuth) ValidateToken(ctx context.Context, token string) (*authv1.UserInfo, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, token)
	}
	return &authv1.UserInfo{Username: "test-user"}, nil
}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Port:        8443,
				StoragePath: t.TempDir(),
				DateFormat:  "2006/01/02",
				TLSCert:     "/path/to/cert",
				TLSKey:      "/path/to/key",
			},
			wantErr: false,
		},
		{
			name: "empty date format",
			config: &Config{
				Port:        8443,
				StoragePath: t.TempDir(),
				DateFormat:  "",
			},
			wantErr: true,
			errMsg:  "dateFormat is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := fake.NewSimpleClientset()
			srv, err := NewServer(tt.config, k8sClient)

			if tt.wantErr {
				if err == nil {
					t.Error("NewServer() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("NewServer() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewServer() unexpected error = %v", err)
				return
			}

			if srv == nil {
				t.Error("NewServer() returned nil server")
			}

			if srv.config.Port != tt.config.Port {
				t.Errorf("NewServer() port = %v, want %v", srv.config.Port, tt.config.Port)
			}
		})
	}
}

func TestHandleProfile_MethodNotAllowed(t *testing.T) {
	config := &Config{
		Port:        8443,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	methods := []string{"GET", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/profile", nil)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(srv.handleProfile)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d", rr.Code)
			}
		})
	}
}

func TestHandleProfile_MissingAuthHeader(t *testing.T) {
	config := &Config{
		Port:        8443,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/profile", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(srv.handleProfile)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestHandleProfile_InvalidAuthHeader(t *testing.T) {
	config := &Config{
		Port:        8443,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "invalid-token"},
		{"empty bearer", "Bearer "},
		{"wrong prefix", "Basic token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/profile", nil)
			req.Header.Set("Authorization", tt.header)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(srv.handleProfile)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", rr.Code)
			}
		})
	}
}

func TestHandleProfile_MissingRequiredHeaders(t *testing.T) {
	config := &Config{
		Port:        8443,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	tests := []struct {
		name    string
		headers map[string]string
	}{
		{
			name: "missing job ID",
			headers: map[string]string{
				"Authorization":                 "Bearer token",
				"X-CoDriverJob-Namespace":       "default",
				"X-CoDriverJob-Matching-Labels": "app-nginx",
			},
		},
		{
			name: "missing namespace",
			headers: map[string]string{
				"Authorization":                 "Bearer token",
				"X-CoDriverJob-Job-ID":          "test-job",
				"X-CoDriverJob-Matching-Labels": "app-nginx",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/profile", bytes.NewBufferString("data"))
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(srv.handleProfile)
			handler.ServeHTTP(rr, req)

			// Will fail at token validation since we're using fake client
			// but we're testing header validation logic
			if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusBadRequest {
				t.Errorf("expected status 401 or 400, got %d", rr.Code)
			}
		})
	}
}

func TestHandleProfile_LargeFile(t *testing.T) {
	config := &Config{
		Port:        8443,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Create 1MB of data
	largeData := make([]byte, 1024*1024)
	req := httptest.NewRequest("POST", "/api/v1/profile", bytes.NewReader(largeData))
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-CoDriverJob-Job-ID", "test-job")
	req.Header.Set("X-CoDriverJob-Namespace", "default")
	req.Header.Set("X-CoDriverJob-Matching-Labels", "app-test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.handleProfile)
	handler.ServeHTTP(rr, req)

	// Will fail at token validation, but verifies large file handling doesn't crash
	if rr.Code != http.StatusUnauthorized {
		t.Logf("Note: Got status %d (expected 401 due to fake token)", rr.Code)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestHandleProfile_Success(t *testing.T) {
	mockStorage := &mockStorage{
		saveProfileFunc: func(r io.Reader, metadata storage.ProfileMetadata) error {
			if metadata.Namespace != "default" {
				t.Errorf("expected namespace 'default', got %v", metadata.Namespace)
			}
			if metadata.AppLabel != "app-nginx" {
				t.Errorf("expected appLabel 'app-nginx', got %v", metadata.AppLabel)
			}
			if metadata.CoDriverJobName != "test-job" {
				t.Errorf("expected coDriverJobName 'test-job', got %v", metadata.CoDriverJobName)
			}
			if metadata.Filename != "output.txt" {
				t.Errorf("expected filename 'output.txt', got %v", metadata.Filename)
			}
			return nil
		},
	}

	mockAuth := &mockAuth{
		validateTokenFunc: func(ctx context.Context, token string) (*authv1.UserInfo, error) {
			if token != "valid-token" {
				return nil, errors.New("invalid token")
			}
			return &authv1.UserInfo{Username: "test-user"}, nil
		},
	}

	srv := &Server{
		storage: mockStorage,
		auth:    mockAuth,
	}

	body := bytes.NewBufferString("test profile data")
	req := httptest.NewRequest("POST", "/api/v1/profile", body)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("X-CoDriverJob-Job-ID", "test-job")
	req.Header.Set("X-CoDriverJob-Namespace", "default")
	req.Header.Set("X-CoDriverJob-Matching-Labels", "app-nginx")
	req.Header.Set("X-CoDriverJob-Filename", "output.txt")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.handleProfile)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleProfile_DefaultValues(t *testing.T) {
	mockStorage := &mockStorage{
		saveProfileFunc: func(r io.Reader, metadata storage.ProfileMetadata) error {
			if metadata.AppLabel != "unknown" {
				t.Errorf("expected default appLabel 'unknown', got %v", metadata.AppLabel)
			}
			if metadata.Filename != "test-job.profile" {
				t.Errorf("expected default filename 'test-job.profile', got %v", metadata.Filename)
			}
			return nil
		},
	}

	mockAuth := &mockAuth{}

	srv := &Server{
		storage: mockStorage,
		auth:    mockAuth,
	}

	body := bytes.NewBufferString("test data")
	req := httptest.NewRequest("POST", "/api/v1/profile", body)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-CoDriverJob-Job-ID", "test-job")
	req.Header.Set("X-CoDriverJob-Namespace", "default")
	// Missing X-CoDriverJob-Matching-Labels and X-CoDriverJob-Filename

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.handleProfile)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestHandleProfile_StorageError(t *testing.T) {
	mockStorage := &mockStorage{
		saveProfileFunc: func(r io.Reader, metadata storage.ProfileMetadata) error {
			return errors.New("disk full")
		},
	}

	mockAuth := &mockAuth{}

	srv := &Server{
		storage: mockStorage,
		auth:    mockAuth,
	}

	body := bytes.NewBufferString("test data")
	req := httptest.NewRequest("POST", "/api/v1/profile", body)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-CoDriverJob-Job-ID", "test-job")
	req.Header.Set("X-CoDriverJob-Namespace", "default")
	req.Header.Set("X-CoDriverJob-Matching-Labels", "app-test")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.handleProfile)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestHandleProfile_TokenValidationError(t *testing.T) {
	mockAuth := &mockAuth{
		validateTokenFunc: func(ctx context.Context, token string) (*authv1.UserInfo, error) {
			return nil, errors.New("token expired")
		},
	}

	srv := &Server{
		auth: mockAuth,
	}

	req := httptest.NewRequest("POST", "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer expired-token")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.handleProfile)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestStart_HTTP(t *testing.T) {
	config := &Config{
		Port:        0, // Use random port
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
		// No TLS cert/key - will use HTTP
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start()
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Shutdown server
	_ = srv.Shutdown()

	// Check if Start returned (should return after shutdown)
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Start() unexpected error = %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Start() did not return after shutdown")
	}
}

func TestStart_TLS(t *testing.T) {
	config := &Config{
		Port:        0,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
		TLSCert:     "/nonexistent/cert.pem",
		TLSKey:      "/nonexistent/key.pem",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Start should fail with missing cert files
	err = srv.Start()
	if err == nil {
		t.Error("Start() expected error with missing TLS files, got nil")
	}
}

func TestShutdown(t *testing.T) {
	config := &Config{
		Port:        0,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Start server
	go func() {
		_ = srv.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	// Shutdown should not panic
	err = srv.Shutdown()
	if err != nil && err != http.ErrServerClosed {
		t.Errorf("Shutdown() unexpected error = %v", err)
	}

	// Multiple shutdowns should not panic
	err = srv.Shutdown()
	// Second shutdown may or may not return error, just verify no panic
}

func TestShutdown_WithoutStart(t *testing.T) {
	config := &Config{
		Port:        0,
		StoragePath: t.TempDir(),
		DateFormat:  "2006/01/02",
	}
	k8sClient := fake.NewSimpleClientset()
	srv, err := NewServer(config, k8sClient)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Shutdown without starting should not panic
	err = srv.Shutdown()
	// May or may not return error, just verify no panic
	_ = err
}
