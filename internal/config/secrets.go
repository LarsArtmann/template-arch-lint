package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SecretsProvider defines the interface for secrets management backends
type SecretsProvider interface {
	GetSecret(ctx context.Context, key string) (string, error)
	SetSecret(ctx context.Context, key, value string) error
	DeleteSecret(ctx context.Context, key string) error
	ListSecrets(ctx context.Context) ([]string, error)
	IsAvailable(ctx context.Context) bool
	Close() error
}

// SecretsManager manages secrets from multiple providers with fallbacks
type SecretsManager struct {
	providers []SecretsProvider
	cache     map[string]secretCacheEntry
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
}

type secretCacheEntry struct {
	value     string
	expiresAt time.Time
}

// SecretConfig contains secrets management configuration
type SecretConfig struct {
	Provider    string        `mapstructure:"provider" validate:"required,oneof=vault kubernetes env file"`
	VaultConfig VaultConfig   `mapstructure:"vault"`
	K8sConfig   K8sConfig     `mapstructure:"kubernetes"`
	FileConfig  FileConfig    `mapstructure:"file"`
	CacheTTL    time.Duration `mapstructure:"cache_ttl"`
}

// VaultConfig contains HashiCorp Vault configuration
type VaultConfig struct {
	Address    string            `mapstructure:"address" validate:"required_if=Provider vault"`
	Token      string            `mapstructure:"token"`
	TokenFile  string            `mapstructure:"token_file"`
	Mount      string            `mapstructure:"mount"`
	Path       string            `mapstructure:"path"`
	Namespace  string            `mapstructure:"namespace"`
	Headers    map[string]string `mapstructure:"headers"`
	TLSSkip    bool              `mapstructure:"tls_skip"`
	TLSCert    string            `mapstructure:"tls_cert"`
	TLSKey     string            `mapstructure:"tls_key"`
	TLSCACert  string            `mapstructure:"tls_ca_cert"`
}

// K8sConfig contains Kubernetes secrets configuration
type K8sConfig struct {
	Namespace       string `mapstructure:"namespace"`
	SecretName      string `mapstructure:"secret_name"`
	KubeconfigPath  string `mapstructure:"kubeconfig_path"`
	InCluster       bool   `mapstructure:"in_cluster"`
	ServiceAccount  string `mapstructure:"service_account"`
}

// FileConfig contains file-based secrets configuration
type FileConfig struct {
	Path      string `mapstructure:"path"`
	Format    string `mapstructure:"format" validate:"oneof=json yaml env"`
	Encrypted bool   `mapstructure:"encrypted"`
	KeyFile   string `mapstructure:"key_file"`
}

// NewSecretsManager creates a new secrets manager with the specified providers
func NewSecretsManager(config SecretConfig) (*SecretsManager, error) {
	sm := &SecretsManager{
		providers: make([]SecretsProvider, 0),
		cache:     make(map[string]secretCacheEntry),
		cacheTTL:  config.CacheTTL,
	}

	if sm.cacheTTL == 0 {
		sm.cacheTTL = 5 * time.Minute
	}

	// Initialize providers based on configuration
	switch config.Provider {
	case "vault":
		provider, err := NewVaultProvider(config.VaultConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Vault provider: %w", err)
		}
		sm.providers = append(sm.providers, provider)

	case "kubernetes":
		provider, err := NewK8sProvider(config.K8sConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes provider: %w", err)
		}
		sm.providers = append(sm.providers, provider)

	case "env":
		sm.providers = append(sm.providers, NewEnvProvider())

	case "file":
		provider, err := NewFileProvider(config.FileConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create file provider: %w", err)
		}
		sm.providers = append(sm.providers, provider)
	}

	// Always add environment variables as fallback
	if config.Provider != "env" {
		sm.providers = append(sm.providers, NewEnvProvider())
	}

	return sm, nil
}

// GetSecret retrieves a secret from the first available provider
func (sm *SecretsManager) GetSecret(ctx context.Context, key string) (string, error) {
	// Check cache first
	if value, found := sm.getCachedSecret(key); found {
		return value, nil
	}

	// Try each provider in order
	for _, provider := range sm.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		value, err := provider.GetSecret(ctx, key)
		if err != nil {
			// Log error and continue to next provider
			fmt.Printf("Provider failed to get secret '%s': %v\\n", key, err)
			continue
		}

		// Cache the result
		sm.cacheSecret(key, value)
		return value, nil
	}

	return "", fmt.Errorf("secret '%s' not found in any provider", key)
}

// SetSecret stores a secret using the first available provider
func (sm *SecretsManager) SetSecret(ctx context.Context, key, value string) error {
	for _, provider := range sm.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		err := provider.SetSecret(ctx, key, value)
		if err != nil {
			fmt.Printf("Provider failed to set secret '%s': %v\\n", key, err)
			continue
		}

		// Update cache
		sm.cacheSecret(key, value)
		return nil
	}

	return fmt.Errorf("failed to set secret '%s' in any provider", key)
}

// DeleteSecret removes a secret from all providers
func (sm *SecretsManager) DeleteSecret(ctx context.Context, key string) error {
	var lastErr error
	deleted := false

	for _, provider := range sm.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}

		err := provider.DeleteSecret(ctx, key)
		if err != nil {
			lastErr = err
			continue
		}
		deleted = true
	}

	if deleted {
		// Remove from cache
		sm.removeCachedSecret(key)
		return nil
	}

	if lastErr != nil {
		return lastErr
	}

	return fmt.Errorf("secret '%s' not found in any provider", key)
}

// getCachedSecret retrieves a secret from cache if not expired
func (sm *SecretsManager) getCachedSecret(key string) (string, bool) {
	sm.cacheMu.RLock()
	defer sm.cacheMu.RUnlock()

	entry, exists := sm.cache[key]
	if !exists || time.Now().After(entry.expiresAt) {
		return "", false
	}

	return entry.value, true
}

// cacheSecret stores a secret in cache
func (sm *SecretsManager) cacheSecret(key, value string) {
	sm.cacheMu.Lock()
	defer sm.cacheMu.Unlock()

	sm.cache[key] = secretCacheEntry{
		value:     value,
		expiresAt: time.Now().Add(sm.cacheTTL),
	}
}

// removeCachedSecret removes a secret from cache
func (sm *SecretsManager) removeCachedSecret(key string) {
	sm.cacheMu.Lock()
	defer sm.cacheMu.Unlock()

	delete(sm.cache, key)
}

// Close closes all providers and cleans up resources
func (sm *SecretsManager) Close() error {
	for _, provider := range sm.providers {
		if err := provider.Close(); err != nil {
			fmt.Printf("Error closing provider: %v\\n", err)
		}
	}

	// Clear cache
	sm.cacheMu.Lock()
	sm.cache = make(map[string]secretCacheEntry)
	sm.cacheMu.Unlock()

	return nil
}

// EnvironmentProvider implements secrets provider for environment variables
type EnvironmentProvider struct{}

// NewEnvProvider creates a new environment variables provider
func NewEnvProvider() *EnvironmentProvider {
	return &EnvironmentProvider{}
}

func (ep *EnvironmentProvider) GetSecret(ctx context.Context, key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable '%s' not found", key)
	}
	return value, nil
}

func (ep *EnvironmentProvider) SetSecret(ctx context.Context, key, value string) error {
	return fmt.Errorf("setting environment variables is not supported")
}

func (ep *EnvironmentProvider) DeleteSecret(ctx context.Context, key string) error {
	return fmt.Errorf("deleting environment variables is not supported")
}

func (ep *EnvironmentProvider) ListSecrets(ctx context.Context) ([]string, error) {
	var secrets []string
	for _, env := range os.Environ() {
		if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
			secrets = append(secrets, parts[0])
		}
	}
	return secrets, nil
}

func (ep *EnvironmentProvider) IsAvailable(ctx context.Context) bool {
	return true
}

func (ep *EnvironmentProvider) Close() error {
	return nil
}

// FileProvider implements secrets provider for file-based secrets
type FileProvider struct {
	config  FileConfig
	secrets map[string]string
	mu      sync.RWMutex
}

// NewFileProvider creates a new file-based secrets provider
func NewFileProvider(config FileConfig) (*FileProvider, error) {
	fp := &FileProvider{
		config:  config,
		secrets: make(map[string]string),
	}

	if err := fp.loadSecrets(); err != nil {
		return nil, fmt.Errorf("failed to load secrets from file: %w", err)
	}

	return fp, nil
}

func (fp *FileProvider) loadSecrets() error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if _, err := os.Stat(fp.config.Path); os.IsNotExist(err) {
		// File doesn't exist, start with empty secrets
		return nil
	}

	data, err := os.ReadFile(fp.config.Path)
	if err != nil {
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	switch fp.config.Format {
	case "json":
		return json.Unmarshal(data, &fp.secrets)
	case "env":
		return fp.parseEnvFormat(string(data))
	default:
		return fmt.Errorf("unsupported file format: %s", fp.config.Format)
	}
}

func (fp *FileProvider) parseEnvFormat(content string) error {
	fp.secrets = make(map[string]string)
	lines := strings.Split(content, "\\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			fp.secrets[parts[0]] = parts[1]
		}
	}
	
	return nil
}

func (fp *FileProvider) saveSecrets() error {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fp.config.Path), 0700); err != nil {
		return fmt.Errorf("failed to create secrets directory: %w", err)
	}

	var data []byte
	var err error

	switch fp.config.Format {
	case "json":
		data, err = json.MarshalIndent(fp.secrets, "", "  ")
	case "env":
		var lines []string
		for key, value := range fp.secrets {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}
		data = []byte(strings.Join(lines, "\\n"))
	default:
		return fmt.Errorf("unsupported file format: %s", fp.config.Format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	return os.WriteFile(fp.config.Path, data, 0600)
}

func (fp *FileProvider) GetSecret(ctx context.Context, key string) (string, error) {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	value, exists := fp.secrets[key]
	if !exists {
		return "", fmt.Errorf("secret '%s' not found", key)
	}

	return value, nil
}

func (fp *FileProvider) SetSecret(ctx context.Context, key, value string) error {
	fp.mu.Lock()
	fp.secrets[key] = value
	fp.mu.Unlock()

	return fp.saveSecrets()
}

func (fp *FileProvider) DeleteSecret(ctx context.Context, key string) error {
	fp.mu.Lock()
	delete(fp.secrets, key)
	fp.mu.Unlock()

	return fp.saveSecrets()
}

func (fp *FileProvider) ListSecrets(ctx context.Context) ([]string, error) {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	var keys []string
	for key := range fp.secrets {
		keys = append(keys, key)
	}

	return keys, nil
}

func (fp *FileProvider) IsAvailable(ctx context.Context) bool {
	return true
}

func (fp *FileProvider) Close() error {
	return nil
}

// VaultProvider implements secrets provider for HashiCorp Vault
type VaultProvider struct {
	config VaultConfig
	client VaultClient // Simplified interface - in real implementation would use official Vault client
}

// VaultClient is a simplified interface for Vault operations
type VaultClient interface {
	Read(path string) (map[string]interface{}, error)
	Write(path string, data map[string]interface{}) error
	Delete(path string) error
	List(path string) ([]string, error)
	IsHealthy() bool
	Close() error
}

// NewVaultProvider creates a new HashiCorp Vault provider
func NewVaultProvider(config VaultConfig) (*VaultProvider, error) {
	// In a real implementation, you would create a proper Vault client here
	// For now, return a mock implementation
	return &VaultProvider{
		config: config,
		client: &MockVaultClient{}, // Replace with real implementation
	}, nil
}

func (vp *VaultProvider) GetSecret(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf("%s/%s", vp.config.Path, key)
	data, err := vp.client.Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read from Vault: %w", err)
	}

	if value, ok := data["value"].(string); ok {
		return value, nil
	}

	return "", fmt.Errorf("secret '%s' not found in Vault", key)
}

func (vp *VaultProvider) SetSecret(ctx context.Context, key, value string) error {
	path := fmt.Sprintf("%s/%s", vp.config.Path, key)
	data := map[string]interface{}{
		"value": value,
	}

	return vp.client.Write(path, data)
}

func (vp *VaultProvider) DeleteSecret(ctx context.Context, key string) error {
	path := fmt.Sprintf("%s/%s", vp.config.Path, key)
	return vp.client.Delete(path)
}

func (vp *VaultProvider) ListSecrets(ctx context.Context) ([]string, error) {
	return vp.client.List(vp.config.Path)
}

func (vp *VaultProvider) IsAvailable(ctx context.Context) bool {
	return vp.client.IsHealthy()
}

func (vp *VaultProvider) Close() error {
	return vp.client.Close()
}

// MockVaultClient is a mock implementation for testing
type MockVaultClient struct {
	data map[string]map[string]interface{}
	mu   sync.RWMutex
}

func (mvc *MockVaultClient) Read(path string) (map[string]interface{}, error) {
	mvc.mu.RLock()
	defer mvc.mu.RUnlock()

	if mvc.data == nil {
		return nil, fmt.Errorf("secret not found")
	}

	data, exists := mvc.data[path]
	if !exists {
		return nil, fmt.Errorf("secret not found")
	}

	return data, nil
}

func (mvc *MockVaultClient) Write(path string, data map[string]interface{}) error {
	mvc.mu.Lock()
	defer mvc.mu.Unlock()

	if mvc.data == nil {
		mvc.data = make(map[string]map[string]interface{})
	}

	mvc.data[path] = data
	return nil
}

func (mvc *MockVaultClient) Delete(path string) error {
	mvc.mu.Lock()
	defer mvc.mu.Unlock()

	if mvc.data != nil {
		delete(mvc.data, path)
	}

	return nil
}

func (mvc *MockVaultClient) List(path string) ([]string, error) {
	mvc.mu.RLock()
	defer mvc.mu.RUnlock()

	var keys []string
	if mvc.data != nil {
		for key := range mvc.data {
			if strings.HasPrefix(key, path) {
				keys = append(keys, key)
			}
		}
	}

	return keys, nil
}

func (mvc *MockVaultClient) IsHealthy() bool {
	return true
}

func (mvc *MockVaultClient) Close() error {
	return nil
}

// K8sProvider implements secrets provider for Kubernetes secrets
type K8sProvider struct {
	config K8sConfig
	client K8sClient // Simplified interface - in real implementation would use client-go
}

// K8sClient is a simplified interface for Kubernetes operations
type K8sClient interface {
	GetSecret(namespace, name, key string) (string, error)
	SetSecret(namespace, name, key, value string) error
	DeleteSecret(namespace, name, key string) error
	ListSecrets(namespace, name string) ([]string, error)
	IsAvailable() bool
	Close() error
}

// NewK8sProvider creates a new Kubernetes secrets provider
func NewK8sProvider(config K8sConfig) (*K8sProvider, error) {
	// In a real implementation, you would create a proper Kubernetes client here
	return &K8sProvider{
		config: config,
		client: &MockK8sClient{}, // Replace with real implementation
	}, nil
}

func (kp *K8sProvider) GetSecret(ctx context.Context, key string) (string, error) {
	return kp.client.GetSecret(kp.config.Namespace, kp.config.SecretName, key)
}

func (kp *K8sProvider) SetSecret(ctx context.Context, key, value string) error {
	return kp.client.SetSecret(kp.config.Namespace, kp.config.SecretName, key, value)
}

func (kp *K8sProvider) DeleteSecret(ctx context.Context, key string) error {
	return kp.client.DeleteSecret(kp.config.Namespace, kp.config.SecretName, key)
}

func (kp *K8sProvider) ListSecrets(ctx context.Context) ([]string, error) {
	return kp.client.ListSecrets(kp.config.Namespace, kp.config.SecretName)
}

func (kp *K8sProvider) IsAvailable(ctx context.Context) bool {
	return kp.client.IsAvailable()
}

func (kp *K8sProvider) Close() error {
	return kp.client.Close()
}

// MockK8sClient is a mock implementation for testing
type MockK8sClient struct {
	secrets map[string]map[string]string
	mu      sync.RWMutex
}

func (mkc *MockK8sClient) GetSecret(namespace, name, key string) (string, error) {
	mkc.mu.RLock()
	defer mkc.mu.RUnlock()

	secretKey := fmt.Sprintf("%s/%s", namespace, name)
	if mkc.secrets == nil {
		return "", fmt.Errorf("secret not found")
	}

	secret, exists := mkc.secrets[secretKey]
	if !exists {
		return "", fmt.Errorf("secret not found")
	}

	value, exists := secret[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	return value, nil
}

func (mkc *MockK8sClient) SetSecret(namespace, name, key, value string) error {
	mkc.mu.Lock()
	defer mkc.mu.Unlock()

	if mkc.secrets == nil {
		mkc.secrets = make(map[string]map[string]string)
	}

	secretKey := fmt.Sprintf("%s/%s", namespace, name)
	if _, exists := mkc.secrets[secretKey]; !exists {
		mkc.secrets[secretKey] = make(map[string]string)
	}

	mkc.secrets[secretKey][key] = value
	return nil
}

func (mkc *MockK8sClient) DeleteSecret(namespace, name, key string) error {
	mkc.mu.Lock()
	defer mkc.mu.Unlock()

	secretKey := fmt.Sprintf("%s/%s", namespace, name)
	if mkc.secrets != nil {
		if secret, exists := mkc.secrets[secretKey]; exists {
			delete(secret, key)
		}
	}

	return nil
}

func (mkc *MockK8sClient) ListSecrets(namespace, name string) ([]string, error) {
	mkc.mu.RLock()
	defer mkc.mu.RUnlock()

	secretKey := fmt.Sprintf("%s/%s", namespace, name)
	if mkc.secrets == nil {
		return nil, nil
	}

	secret, exists := mkc.secrets[secretKey]
	if !exists {
		return nil, nil
	}

	var keys []string
	for key := range secret {
		keys = append(keys, key)
	}

	return keys, nil
}

func (mkc *MockK8sClient) IsAvailable() bool {
	return true
}

func (mkc *MockK8sClient) Close() error {
	return nil
}

// ExpandSecrets expands environment variable placeholders in configuration with actual secrets
func ExpandSecrets(config *Config, secretsManager *SecretsManager) error {
	ctx := context.Background()
	
	// Expand database DSN
	if dsn, err := expandSecret(config.Database.DSN, secretsManager, ctx); err == nil {
		config.Database.DSN = dsn
	}
	
	// Expand observability endpoints and tokens
	if endpoint, err := expandSecret(config.Observability.Tracing.Endpoint, secretsManager, ctx); err == nil {
		config.Observability.Tracing.Endpoint = endpoint
	}
	
	if endpoint, err := expandSecret(config.Observability.Metrics.Endpoint, secretsManager, ctx); err == nil {
		config.Observability.Metrics.Endpoint = endpoint
	}
	
	// Expand OTLP headers
	for key, value := range config.Observability.Exporters.OTLP.Headers {
		if expanded, err := expandSecret(value, secretsManager, ctx); err == nil {
			config.Observability.Exporters.OTLP.Headers[key] = expanded
		}
	}
	
	// Expand cache Redis URL
	if redisURL, err := expandSecret(config.Cache.RedisURL, secretsManager, ctx); err == nil {
		config.Cache.RedisURL = redisURL
	}
	
	// Expand TLS certificate paths (could be secret references)
	if certFile, err := expandSecret(config.Security.TLS.CertFile, secretsManager, ctx); err == nil {
		config.Security.TLS.CertFile = certFile
	}
	
	if keyFile, err := expandSecret(config.Security.TLS.KeyFile, secretsManager, ctx); err == nil {
		config.Security.TLS.KeyFile = keyFile
	}
	
	return nil
}

// expandSecret expands a single secret reference if it starts with ${}
func expandSecret(value string, secretsManager *SecretsManager, ctx context.Context) (string, error) {
	if !strings.HasPrefix(value, "${") || !strings.HasSuffix(value, "}") {
		return value, fmt.Errorf("not a secret reference")
	}
	
	secretKey := strings.TrimPrefix(strings.TrimSuffix(value, "}"), "${")
	return secretsManager.GetSecret(ctx, secretKey)
}