package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nikole-dunixi/add-direnv-gotool/internal/direnv"
	"github.com/nikole-dunixi/add-direnv-gotool/internal/gocmd"
	flag "github.com/spf13/pflag"
	"golang.org/x/mod/module"
)

const (
	defaultSubdirectoryName string = ".gotools"
)

type Settings struct {
	// The name of the binary being executed
	BinaryName string
	// The path to create new files
	MainDirectory string
	// The path to the .envrc file
	EnvRCPath string
	// The subdirectory name to use when creating relative files
	ToolsSubdirectoryName string
	// The package path to use with `go get -tool`
	PackagePath string
	// The command that will be used to reference
	CommandName string
	// Indicate we will need to create a dedicated module file
	IsolateModule bool
}

type ArgumentError struct {
	Message string
}

func (err ArgumentError) Error() string {
	return err.Message
}

func ParseArgs(ctx context.Context, executable string, args ...string) (Settings, error) {
	var toolsSubdirectory string
	var commandName string
	var isolateModule bool
	binaryName := filepath.Base(executable)

	flags := flag.NewFlagSet(executable, flag.ExitOnError)
	flags.StringVarP(&toolsSubdirectory, "tools-directory-name", "d", defaultSubdirectoryName, "Specify the subdirectory to create resources under")
	flags.MarkHidden("tools-directory-name")
	flags.StringVar(&commandName, "command-name", "", "")
	flags.BoolVar(&isolateModule, "isolate-module", false, "Create a dedicated modfile for the 'go tool' to use")
	flags.Usage = func() {

		println(fmt.Sprintf("Usage of %s:", binaryName))
		println()
		println(fmt.Sprintf("\t%s [--isolate-module] [--command-name=value] [package-path]", binaryName))
		println()
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return Settings{}, fmt.Errorf("could not parse cli arguments: %w", err)
	}
	if flags.NArg() != 1 {
		return Settings{}, ArgumentError{
			Message: "a single package path must be provided",
		}
	}

	if toolsSubdirectory == "" {
		toolsSubdirectory = defaultSubdirectoryName
	}

	envrcPath, err := direnv.FindEnvrc(ctx)
	if errors.Is(err, direnv.ErrNotDirenvManaged) {
		return Settings{}, ArgumentError{
			Message: "tool must be executed in project managed by direnv",
		}
	} else if err != nil {
		return Settings{}, fmt.Errorf("something went wrong while attempting to interface with direnv: %w", err)
	}
	mainDirectory := filepath.Dir(envrcPath)

	packagePath := strings.TrimSpace(flags.Arg(0))
	if commandName == "" {
		tmpCommandName, ok := DetermineCommandName(packagePath)
		if !ok {
			return Settings{}, ArgumentError{
				Message: "could not determine command-name from package path, adjust the package path or use the --command-name flag",
			}
		}
		commandName = tmpCommandName
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		return Settings{}, fmt.Errorf("could not get current directory: %w", err)
	}

	if isGoModule, err := gocmd.CheckInModule(ctx, currentDirectory); err != nil {
		return Settings{}, fmt.Errorf("could not determine if current directory is within a go module: %w", err)
	} else if !isGoModule {
		slog.DebugContext(ctx, "project is not a golang module, forcing isolated module usage")
		isolateModule = true
	}

	settings := Settings{
		BinaryName:            binaryName,
		MainDirectory:         mainDirectory,
		EnvRCPath:             envrcPath,
		ToolsSubdirectoryName: toolsSubdirectory,
		PackagePath:           packagePath,
		CommandName:           commandName,
		IsolateModule:         isolateModule,
	}

	return settings, nil
}

func DetermineCommandName(packagePath string) (string, bool) {
	pathWithoutVersion, _, ok := module.SplitPathVersion(strings.TrimSpace(packagePath))
	if !ok {
		return "", false
	}
	commandName := path.Base(pathWithoutVersion)
	if index := strings.Index(commandName, "@"); index >= 0 {
		commandName = commandName[:index]
	}
	return commandName, commandName != ""
}
