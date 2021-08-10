package testhelper

import (
	"context"
	"fmt"

	myconfig "github.com/Fox520/away_backend/config"
	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateTestContainer(ctx context.Context, cfg myconfig.Config) (testcontainers.Container, string, error) {
	var env = map[string]string{
		"POSTGRES_PASSWORD": cfg.DBPassword,
		"POSTGRES_USER":     "postgres",
		"POSTGRES_DB":       cfg.DBName,
	}
	var port = "5432/tcp"
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:%s@localhost:%s/%s?sslmode=disable", cfg.DBPassword, port.Port(), cfg.DBName)
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:latest",
			ExposedPorts: []string{port},
			Cmd:          []string{"postgres", "-c", "fsync=off"},
			Env:          env,
			WaitingFor:   wait.ForSQL(nat.Port(port), "postgres", dbURL),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, "", fmt.Errorf("failed to start container: %s", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return container, "", fmt.Errorf("failed to get container external port: %s", err)
	}

	return container, mappedPort.Port(), nil
}
