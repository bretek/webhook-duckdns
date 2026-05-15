package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(GroupName,
		&duckDNSProviderSolver{},
	)
}

type duckDNSProviderSolver struct {
	logger *slog.Logger
	client *kubernetes.Clientset
	ctx    context.Context
}

type duckDNSProviderConfig struct {
	APIKeySecretRef corev1.SecretKeySelector `json:"apiKeySecretRef"`
}

func (c *duckDNSProviderSolver) Name() string {
	return "duckdns"
}

func (c *duckDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		c.logger.Error("Bad config", "error", err)
		return err
	}

	apiKey, err := c.getApiKey(ch.ResourceNamespace, cfg.APIKeySecretRef.Name, cfg.APIKeySecretRef.Key)
	if err != nil {
		return err
	}

	domain := strings.TrimSuffix(ch.ResolvedZone, ".")
	url := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&txt=%s", domain, *apiKey, ch.Key)
	c.logger.Debug("Updating TXT...", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		body := string(bodyBytes)
		if body == "OK" {
			c.logger.Info("Updated successfully", "response", body)
			return nil
		}
		c.logger.Error("Bad response", "body", body)
		return errors.New("Failed to update TXT")
	} else {
		c.logger.Error("Bad response", "status", resp.Status)
		return errors.New("Failed to update TXT")
	}
}

func (c *duckDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	apiKey, err := c.getApiKey(ch.ResourceNamespace, cfg.APIKeySecretRef.Name, cfg.APIKeySecretRef.Key)
	if err != nil {
		return err
	}

	domain := strings.TrimSuffix(ch.ResolvedZone, ".")
	url := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&txt=&clear=true", domain, *apiKey)
	c.logger.Debug("Clearing TXT...", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		body := string(bodyBytes)
		if body == "OK" {
			c.logger.Info("Cleared successfully", "response", body)
			return nil
		}
		c.logger.Error("Bad response", "body", body)
		return errors.New("Failed to clear TXT")
	} else {
		c.logger.Error("Bad response", "status", resp.Status)
		return errors.New("Failed to clear TXT")
	}
}

func (c *duckDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: getLogLevelFromEnv()}))
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.logger = logger
	c.client = cl
	c.ctx = context.Background()

	return nil
}

func (c *duckDNSProviderSolver) getApiKey(namespace, secretName, key string) (*string, error) {
	secret, err := c.client.CoreV1().Secrets(namespace).Get(c.ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Kubernetes secret %s/%s not found: %v", namespace, secretName, err)
	}

	apiKeyBinary, ok := secret.Data[key]
	if !ok {
		return nil, fmt.Errorf("Key `%q` not found in secret %s/%s", key, namespace, secretName)
	}
	apiKey := strings.TrimSpace(string(apiKeyBinary))
	c.logger.Info("Secret loaded", "secret name", secret.Name)
	return &apiKey, nil
}

func getLogLevelFromEnv() slog.Level {
	levelStr := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn

	}
}

func loadConfig(cfgJSON *extapi.JSON) (duckDNSProviderConfig, error) {
	cfg := duckDNSProviderConfig{}
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
