package testhelper

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateTestContainer(ctx context.Context) (testcontainers.Container, string, error) {
	var env = map[string]string{
		"POSTGRES_PASSWORD": "secret",
		"POSTGRES_USER":     "root",
		"POSTGRES_DB":       "away",
	}
	var port = "5432/tcp"
	dbURL := func(port nat.Port) string {
		return fmt.Sprintf("postgres://root:%s@localhost:%s/%s?sslmode=disable", "secret", port.Port(), "away")
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "postgres:14-alpine",
			ExposedPorts:    []string{port},
			Cmd:             []string{"postgres", "-c", "fsync=off"},
			Env:             env,
			WaitingFor:      wait.ForSQL(nat.Port(port), "postgres", dbURL),
			AlwaysPullImage: false,
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
