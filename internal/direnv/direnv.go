package direnv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
)

var ErrNotDirenvManaged error = errors.New("directory is not managed by direnv")

type CommandStatus struct {
	Config Config `json:"config"`
	State  State  `json:"state"`
}

type Config struct {
	ConfigDir string `json:"ConfigDir"`
	SelfPath  string `json:"SelfPath"`
}

type State struct {
	FoundRC  *RC `json:"foundRC"`
	LoadedRC *RC `json:"loadedRC"`
}

type RC struct {
	Allowed int    `json:"allowed"`
	Path    string `json:"path"`
}

func FindEnvrc(ctx context.Context) (string, error) {
	status, err := Status(ctx)
	if err != nil {
		return "", err
	}
	if status.State.FoundRC == nil {
		return "", ErrNotDirenvManaged
	}
	foundRCPath := status.State.FoundRC.Path
	if foundRCPath == "" {
		return "", ErrNotDirenvManaged
	}
	return foundRCPath, nil
}

func Status(ctx context.Context) (CommandStatus, error) {
	result := CommandStatus{}

	buffer := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "direnv", "status", "--json")
	cmd.Stdout = buffer
	if err := cmd.Run(); err != nil {
		return result, err
	}

	if err := json.Unmarshal(buffer.Bytes(), &result); err != nil {
		return result, err
	}

	return result, nil
}
