# add-direnv-gotool

A helper to integrate golang tools into code-bases that
use `direnv`. Naturally this will only work if your project is managed by `direnv`. This does not require the project to be a golang project, but does assume the golang toolchain is installed on the development machine.

See: https://direnv.net/

## What Problem Does this Solve?

Disparate developers each configure and manage their system packages independantly. `direnv` automates configuring project variables. With `go tool`s (introduced in Golang 1.24), one can manage tools as part of a Golang project. `add-direnv-gotool` ties into `direnv` and uses `go tool` to install tools in a consistent way.

For golang projects, the convienient byproduct for cli-tools that are related to a projects dependencies can be kept up to date as simply as running `go get -u ./...`.

For example, Ginkgo is a BDD framework and a cli tool. When the versions are not synchronized, tests will warn that unexpected behavior may occur. Keeping these two components in parity is now trivial. Conversely, it is now equally trivial add a `go tool` that is isolated from the project module when necessary.

### Alternatives to Consider

There are alternative solutions with their own trade-offs to consider.
* [nix-flakes](https://nixos.wiki/wiki/flakes)
* [vscode dev-containers](https://code.visualstudio.com/docs/devcontainers/create-dev-container)

## Installation

To install the application as an independant CLI tool:
```bash
$ go install github.com/nikole-dunixi/add-direnv-gotool@latest
```

See also, [bootstrapping](#bootstrapping)

## Usage

### Adding Tools to a Project

When you execute `add-direnv-gotool`, it will:
* Locate your `.envrc` file
* Create a `.gotools` sub-directory adjacent to `.envrc`
* Append `PATH_add .gotools` to .`envrc` if absent

#### Basic Usage
Navigate to your direnv enabled project and execute similarly to `go get -tool`.

```bash
$ cd your-project-here
$ add-direnv-gotool github.com/onsi/ginkgo/v2/ginkgo@latest
```

A `.gotools` subdirectory will be created, and an executable script with the name `ginkgo` will also be added. The script delegates execution to `go tool`.

```bash
$ ls .gotools/
ginkgo
```

#### Non-Integrated Tools
When you perform a `go get -tool`, a tool directive is added to the project's `go.mod` file. This may be undesireable when the project's dependencies should not be affected by the tool's dependencies. Use the `--isolate-module` flag to create a discrete and dedicated module file.

```bash
$ cd your-project-here
$ add-direnv-gotool --isolate-module golang.org/x/tools/cmd/goimports@latest
```

Similar to the basic usage, this will create a script by the same name as the binary. An additional module and sum file will also be created. The script will reference the module file.

```bash
$ ls .gotools/
goimports
goimports.mod
goimports.sum
```

> Notice:
>
> When using `add-direnv-gotool` in a project that is not a golang module, isolated-module will be assumed and discrete module files will be created to ensure correct functionality despite the absense of a primary go-module file. This still requires a golang toolchain be installed regardless.

#### Bootstrapping
Instead of a global installation, you can install `add-direnv-gotool` on a per-project basis with itself.

```bash
$ go run github.com/nikole-dunixi/add-direnv-gotool@latest \
  --isolate-module github.com/nikole-dunixi/add-direnv-gotool@latest
```

### Examples
For a given tool, the primary choice to make is if you want the depenency to be managed independantly or not.

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [templ](https://templ.guide/) | No | The code used to render templates is from the same package as the CLI tool. It is ideal to keep these in sync. |
```bash
$ add-direnv-gotool github.com/a-h/templ/cmd/templ@latest
```

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [templui](https://templui.io/) | Yes | Unlike templ, templui generates non-dependant code. Since it is not directly intregrated in the final binary, its dependencies should be separate. |
```bash
$ add-direnv-gotool --isolate-module github.com/templui/templui/cmd/templui@latest
```

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [ginkgo](https://onsi.github.io/ginkgo/) | No | The tests that are written using ginkgo's library expect the CLI to match. Mismatches cause tests to emit a warning for potential instability |
```bash
$ add-direnv-gotool github.com/onsi/ginkgo/v2/ginkgo@latest
```

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [mage](https://magefile.org/) | No | Mage is a replacement for `make` which is written directly in golang code in your project. Your module file will track mage by default, thus it is ideal to keep the CLI in sync. |
```bash
$ add-direnv-gotool github.com/magefile/mage@latest
```

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [gow](https://github.com/mitranim/gow) | Yes | Gow is a CLI tool that watches your project for filesystem changes. Since your code never integrates gow as a library, the CLI tool should not have side-effects on your dependencies. |
```bash
$ add-direnv-gotool --isolate-module github.com/mitranim/gow@latest
```

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) | Yes | This tool arranges `import`s but doesn't actually get used in your project. It should be separate. |
```bash
$ add-direnv-gotool --isolate-module golang.org/x/tools/cmd/goimports@latest
```

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [task](https://taskfile.dev/) | Yes | Task is similar to mage as a `make` replacement, but describes actions in a yaml file instead of golang code. It should not impact your project's dependencies. |
```bash
$ add-direnv-gotool --isolate-module github.com/go-task/task/v3/cmd/task@latest
```

| Tool | Should Isolate? | Explaination |
| -- | -- | -- |
| [golangci-lint](https://golangci-lint.run/) | Yes | GolangCI-Lint [expliticly discourages using go-tool](https://golangci-lint.run/docs/welcome/install/local/#install-from-sources). If you decide to use it with go-tool, it should have discrete dependencies for stability. |
```bash
$ add-direnv-gotool --isolate-module github.com/golangci/golangci-lint/v2/cmd/golangci-lint
```
