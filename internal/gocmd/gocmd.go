package gocmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GoGetToolOpts struct {
	PackagePath    string
	ModuleFilepath string
}

func CheckInModule(ctx context.Context, dir string) (bool, error) {
	buffer := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "go", "mod", "edit", "-json")
	cmd.Dir = dir
	cmd.Stdout = buffer
	slog.DebugContext(ctx, "check in module",
		slog.String("cmd", strings.Join(cmd.Args, " ")),
	)
	if err := cmd.Run(); err != nil {
		return false, err
	}
	type goModInfo struct {
		Module struct {
			Path string `json:"Path"`
		} `json:"Module"`
	}
	var moduleInfo goModInfo
	if err := json.Unmarshal(buffer.Bytes(), &moduleInfo); err != nil {
		return false, err
	}
	modulePath := strings.TrimSpace(moduleInfo.Module.Path)
	return modulePath != "", nil
}

func CreateIsolatedModule(ctx context.Context, moduleFilename, moduleName string) error {
	slogattr := slog.Group("module",
		slog.String("filename", moduleFilename),
		slog.String("name", moduleName),
	)
	slog.DebugContext(ctx, "attempting to create isolated module",
		slogattr,
	)
	if stat, err := os.Stat(moduleFilename); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not read filesystem for stat: %w", err)
	} else if stat != nil {
		slog.DebugContext(ctx, "isolated module already exists",
			slog.String("moduleFilename", moduleFilename),
			slog.String("moduleName", moduleName),
		)
		return nil
	}
	cmd := exec.CommandContext(ctx,
		"go", "mod", "init", "--modfile="+moduleFilename, moduleName,
	)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	slog.DebugContext(ctx, "executing command",
		slogattr,
		slog.String("cmd", strings.Join(cmd.Args, " ")),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s", stderr.String())
	}
	return nil
}

func GoGetTool(ctx context.Context, opts GoGetToolOpts) error {
	slog.DebugContext(ctx, "using go get to capture tool")
	args := []string{"get", "-tool"}
	if opts.ModuleFilepath != "" {
		args = append(args, "--modfile="+opts.ModuleFilepath)
	}
	args = append(args, opts.PackagePath)
	cmd := exec.CommandContext(ctx,
		"go", args...,
	)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	slog.DebugContext(ctx, "executing command",
		slog.Group("opts",
			slog.String("ModuleFilepath", opts.ModuleFilepath),
			slog.String("PackagePath", opts.PackagePath),
		),
		slog.String("cmd", strings.Join(cmd.Args, " ")),
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s", stderr.String())
	}
	return nil
}

func ModFilename(dir, moduleName string) (string, error) {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return "", errors.New("module name cannot be empty")
	}

	moduleExtension := ".mod"
	actualExtension := filepath.Ext(moduleName)
	if !strings.EqualFold(actualExtension, moduleExtension) {
		moduleName = moduleName + moduleExtension
	}
	moduleFilePath := filepath.Join(dir, moduleName)
	return moduleFilePath, nil
}
