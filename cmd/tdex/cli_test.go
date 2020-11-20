package main_test

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const password = "hodlhodlhodl"

func TestGenSeed(t *testing.T) {
	container := runNewContainer(t)
	defer stopAndDeleteContainer(container)

	t.Run("should return a new mnemonic", func(t *testing.T) {
		seed, err := runCLICommand(container, "genseed")
		assert.Nil(t, err)
		assert.Equal(t, len(strings.Split(seed, " ")), 24)
	})
}

func TestInitWallet(t *testing.T) {
	container := runNewContainer(t)
	defer stopAndDeleteContainer(container)

	seed, err := runCLICommand(container, "genseed")
	if err != nil {
		t.Error(err)
	}

	t.Run("should init the wallet", func(t *testing.T) {
		_, err := runCLICommand(container, "init", "--seed", seed, "--password", password)
		assert.Nil(t, err)
	})
}

func TestUnlockWallet(t *testing.T) {
	container := runNewContainer(t)
	defer stopAndDeleteContainer(container)

	seed, err := runCLICommand(container, "genseed")
	if err != nil {
		t.Error(err)
	}

	_, err = runCLICommand(container, "init", "--seed", seed, "--password", password)
	if err != nil {
		t.Error(err)
	}

	t.Run("should not return error if password is ok", func(t *testing.T) {
		_, err := runCLICommand(container, "unlock", "--password", password)
		assert.Nil(t, err)
	})
}

func execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := out.String()
	if err == nil {
		return result, nil
	}

	return out.String(), errors.New(fmt.Sprint(err) + ": " + stderr.String())
}

func runNewContainer(t *testing.T) string {
	id := uuid.New().String()
	err := runDaemon(id)
	if err != nil {
		t.Error(err)
	}

	return id
}

func runCLICommand(containerName string, cliCommand string, args ...string) (string, error) {
	commandArgs := []string{"exec", containerName, "tdex", cliCommand}
	commandArgs = append(commandArgs, args...)
	output, err := execute("docker", commandArgs...)
	return output, err
}

func runDaemon(containerName string) error {
	_, err := execute(
		"docker", "run", "--name", containerName,
		// "-p", "9945:9945", "-p", "9000:9000",
		"-d",
		"-e", "TDEX_NETWORK=regtest",
		"-e", "TDEX_EXPLORER_ENDPOINT=http://127.0.0.1:3001",
		"-e", "TDEX_FEE_ACCOUNT_BALANCE_TRESHOLD=1000",
		"-e", "TDEX_BASE_ASSET=5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
		"-e", "TDEX_LOG_LEVEL=5",
		"tdexd:latest",
	)

	return err
}

func stopAndDeleteContainer(containerName string) {
	_, err := execute("docker", "stop", containerName)
	if err != nil {
		panic(err)
	}

	_, err = execute("docker", "container", "rm", containerName)
	if err != nil {
		panic(err)
	}
}
