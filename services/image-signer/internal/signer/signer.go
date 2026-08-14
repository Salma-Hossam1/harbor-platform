package signer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Signer struct {
	cosignBinary string
	privateKey   string
	password     string

	registry         string
	username         string
	registryPassword string

	allowHTTP  bool
	ignoreTLog bool

	allowInsecure bool
}

type HarborEvent struct {
	Type      string `json:"type"`
	EventData struct {
		Resources []struct {
			Digest      string `json:"digest"`
			Tag         string `json:"tag"`
			ResourceURL string `json:"resource_url"`
		} `json:"resources"`

		Repository struct {
			Name         string `json:"name"`
			Namespace    string `json:"namespace"`
			RepoFullName string `json:"repo_full_name"`
		} `json:"repository"`
	} `json:"event_data"`
}

func New(
	cosignBinary string,
	privateKey string,
	password string,
	registry string,
	username string,
	registryPassword string,
	allowHTTP bool,
	ignoreTLog bool,
	allowInsecure bool,
) *Signer {
	return &Signer{
		cosignBinary:     cosignBinary,
		privateKey:       privateKey,
		password:         password,
		registry:         strings.TrimRight(registry, "/"),
		username:         username,
		registryPassword: registryPassword,
		allowHTTP:        allowHTTP,
		ignoreTLog:       ignoreTLog,
		allowInsecure: allowInsecure,
	}
}

func (s *Signer) HandleEvent(ctx context.Context, payload []byte) error {
	var event HarborEvent

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("invalid Harbor event JSON: %w", err)
	}

	if event.Type != "PUSH_ARTIFACT" {
		return nil
	}

	if len(event.EventData.Resources) == 0 {
		return fmt.Errorf("PUSH_ARTIFACT event contains no resources")
	}

	repository := event.EventData.Repository.RepoFullName

	if repository == "" {
		return fmt.Errorf("unable to determine repository")
	}

	for _, resource := range event.EventData.Resources {
		if resource.Digest == "" {
			return fmt.Errorf("resource is missing digest")
		}

		image := fmt.Sprintf(
			"%s/%s@%s",
			s.registry,
			repository,
			resource.Digest,
		)

		fmt.Printf("signing image: %s\n", image)

		if err := s.sign(ctx, image); err != nil {
			return fmt.Errorf("failed to sign %s: %w", image, err)
		}
	}

	return nil
}

func (s *Signer) configureRegistryAuth() (func(), error) {
	auth := base64.StdEncoding.EncodeToString(
		[]byte(s.username + ":" + s.registryPassword),
	)

	dir, err := os.MkdirTemp("", "image-signer-docker-config-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary docker config: %w", err)
	}

	configPath := filepath.Join(dir, "config.json")

	config := fmt.Sprintf(`{
  "auths": {
    "%s": {
      "auth": "%s"
    }
  }
}`, s.registry, auth)

	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("write docker config: %w", err)
	}

	oldDockerConfig := os.Getenv("DOCKER_CONFIG")

	if err := os.Setenv("DOCKER_CONFIG", dir); err != nil {
		_ = os.RemoveAll(dir)
		return nil, fmt.Errorf("set DOCKER_CONFIG: %w", err)
	}

	cleanup := func() {
		if oldDockerConfig == "" {
			_ = os.Unsetenv("DOCKER_CONFIG")
		} else {
			_ = os.Setenv("DOCKER_CONFIG", oldDockerConfig)
		}

		_ = os.RemoveAll(dir)
	}

	return cleanup, nil
}

func (s *Signer) sign(ctx context.Context, image string) error {
	cleanup, err := s.configureRegistryAuth()
	if err != nil {
		return err
	}
	defer cleanup()

	args := []string{
		"sign",
		"--yes",
		"--key",
		s.privateKey,
	}

	if s.allowHTTP {
		args = append(args, "--allow-http-registry")
	}

	if s.allowInsecure {
		args = append(args, "--allow-insecure-registry")
	}

	if s.ignoreTLog {
		args = append(args, "--tlog-upload=false")
	}

	args = append(args, image)

	cmd := exec.CommandContext(
		ctx,
		s.cosignBinary,
		args...,
	)

	cmd.Env = cmd.Environ()

	cmd.Env = append(
		cmd.Env,
		"DOCKER_CONFIG="+os.Getenv("DOCKER_CONFIG"),
		"COSIGN_PASSWORD="+s.password,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"cosign failed: %w: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	fmt.Printf(
		"cosign output: %s\n",
		strings.TrimSpace(string(output)),
	)

	return nil
}
