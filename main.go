package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nikole-dunixi/add-direnv-gotool/internal/assets"
	"github.com/nikole-dunixi/add-direnv-gotool/internal/cli"
	"github.com/nikole-dunixi/add-direnv-gotool/internal/gocmd"
	"github.com/nikole-dunixi/add-direnv-gotool/internal/sloghelper"
)

const (
	UnixDirectoryPermissions      os.FileMode = 0755
	UnixExecutableFilePermissions os.FileMode = 0755
	UnixFilePermissions           os.FileMode = 0644
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	settings, err := cli.ParseArgs(ctx, os.Args[0], os.Args[1:]...)
	if argErr := (cli.ArgumentError{}); errors.As(err, &argErr) {
		sloghelper.FatalContext(ctx, argErr.Message)
	} else if err != nil {
		sloghelper.FatalContextErr(ctx, err, "a problem occurred while handling cli parsing and validation")
	}
	slog.Debug("operation settings determined",
		slog.Group("Settings",
			slog.String("BinaryName", settings.BinaryName),
			slog.String("MainDirectory", settings.MainDirectory),
			slog.String("EnvRCPath", settings.EnvRCPath),
			slog.String("PackagePath", settings.PackagePath),
			slog.String("CommandName", settings.CommandName),
			slog.Bool("IsolateModule", settings.IsolateModule),
		),
	)

	packagePath := settings.PackagePath
	commandName := settings.CommandName
	isolateModule := settings.IsolateModule

	goGetToolOpts := gocmd.GoGetToolOpts{
		PackagePath: packagePath,
	}

	toolsDirectory := filepath.Join(settings.MainDirectory, settings.ToolsSubdirectoryName)
	if err := os.MkdirAll(toolsDirectory, UnixDirectoryPermissions); err != nil {
		sloghelper.FatalContext(ctx, "could not create go tools directory",
			slog.Any("err", err),
			slog.String("toolsDirectory", toolsDirectory),
		)
	}

	if isolateModule {
		moduleFilepath, err := gocmd.ModFilename(toolsDirectory, commandName)
		if err != nil {
			sloghelper.FatalContextErr(ctx, err, "could not determine module name")
		}
		goGetToolOpts.ModuleFilepath = moduleFilepath
		err = gocmd.CreateIsolatedModule(ctx, moduleFilepath, commandName)
		if err != nil {
			sloghelper.FatalContextErr(ctx, err, "could not create isolated go tool module")
		}
	}
	if err = gocmd.GoGetTool(ctx, goGetToolOpts); err != nil {
		sloghelper.FatalContextErr(ctx, err, "could not `go get -tool` for isolated module")
	}

	if err := HandleEnvRC(
		settings.EnvRCPath,
		settings.BinaryName,
		settings.ToolsSubdirectoryName,
	); err != nil {
		sloghelper.FatalContextErr(ctx, err, "could not handle .envrc")
	}

	scriptFilename := filepath.Join(toolsDirectory, commandName)
	if err := HandleScriptGeneration(
		scriptFilename,
		settings.PackagePath,
		settings.CommandName,
		settings.IsolateModule,
	); err != nil {
		sloghelper.FatalContextErr(ctx, err, "could not handle script generation")
	}

	slog.InfoContext(ctx, "success")
}

func HandleEnvRC(envrcPath, binaryName, targetDirectory string) error {
	targetLine := "PATH_add " + targetDirectory

	// Open the file for reading and appending
	f, err := os.OpenFile(envrcPath, os.O_APPEND|os.O_RDWR, UnixFilePermissions)
	if err != nil {
		return fmt.Errorf("could not open envrc for appending: %w", err)
	}
	closeFile := sync.OnceFunc(func() {
		f.Close()
	})
	defer closeFile()

	// Iterate over the file to determine if the line
	// is already present and if not, if the file ends
	// with a newline character already
	prefix := ""
	reader := bufio.NewReader(f)
	isReadingFile := true
	for isReadingFile {
		line, err := reader.ReadString('\n')
		if errors.Is(err, io.EOF) {
			isReadingFile = false
		} else if err != nil {
			return fmt.Errorf("could not read through envrc: %w", err)
		}
		if strings.TrimRight(line, "\n") == targetLine {
			// The file already contains the line. No operation is needed.
			return nil
		} else if endsWithNL, eofWithEmptyLine := strings.HasSuffix(line, "\n"), !isReadingFile && line == ""; !endsWithNL && !eofWithEmptyLine {
			prefix = "\n"
		} else {
			prefix = ""
		}
	}
	// Append the PATH_add and include a prefixed newline
	// if needed.
	_, err = f.WriteString(
		prefix +
			"# Path added by " + binaryName + "\n" +
			"PATH_add " + targetDirectory + "\n",
	)
	if err != nil {
		return fmt.Errorf("could not append PATH_add to direnv .envrc file: %w", err)
	}
	return nil
}

func HandleScriptGeneration(
	scriptFilename,
	packagePath,
	commandName string,
	isolateModule bool,
) error {
	scriptFile, err := os.OpenFile(scriptFilename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, UnixExecutableFilePermissions)
	if err != nil {
		return fmt.Errorf("could not open script file: %w", err)
	}
	defer scriptFile.Close()
	if err := assets.WriteScriptContents(scriptFile, assets.ScriptArgs{
		PackagePath:   packagePath,
		CommandName:   commandName,
		IsolateModule: isolateModule,
	}); err != nil {
		return fmt.Errorf("could not write to script file: %w", err)
	}
	return nil
}
