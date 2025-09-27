package mishan

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stepinto/mishan/gollm"
)

type cobraTool struct {
	cmd *cobra.Command
}

var _ Tool = &cobraTool{}

func NewCobraTool(cmd *cobra.Command) Tool {
	return &cobraTool{cmd: cmd}
}

func (t *cobraTool) Name() string {
	return t.cmd.Name()
}

func (t *cobraTool) Description() string {
	return fmt.Sprintf(`Executes commands from the %s CLI application. Use this tool to interact with the application's functionality through its command-line interface.

Available commands and their usage can be discovered by examining the command structure and help text. The tool accepts command arguments as a string array and executes them against the root command.

Examples:
- To list tasks: ["list"]
- To add a task: ["add", "my task description"]
- To complete a task: ["complete", "1"]

Note: This tool does not support interactive commands that require user input during execution.`, t.cmd.Name())
}

func (t *cobraTool) FunctionDefinition() *gollm.FunctionDefinition {
	return &gollm.FunctionDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &gollm.Schema{
			Type: gollm.TypeObject,
			Properties: map[string]*gollm.Schema{
				"args": {
					Type:        gollm.TypeArray,
					Description: `Array of command arguments to execute. The first element should be the subcommand name, followed by any additional arguments.`,
					Items: &gollm.Schema{
						Type: gollm.TypeString,
					},
				},
			},
			Required: []string{"args"},
		},
	}
}

func (t *cobraTool) Run(ctx context.Context, args map[string]any) (any, error) {
	cmd := t.cmd
	cmdArgs := []string{}
	for _, arg := range args["args"].([]any) {
		cmdArgs = append(cmdArgs, arg.(string))
	}
	cmd.SetArgs(cmdArgs)
	inBuf := bytes.Buffer{}
	outBuf := bytes.Buffer{}
	cmd.SetIn(&inBuf)
	cmd.SetOut(&outBuf)
	err := cmd.ExecuteContext(ctx)
	result := map[string]any{
		"command": strings.Join(cmdArgs, " "),
		"stdout":  outBuf.String(),
	}
	if err != nil {
		result["error"] = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitErr.ExitCode()
		}
	} else {
		result["exit_code"] = 0
	}
	return result, nil
}

func (t *cobraTool) IsInteractive(args map[string]any) (bool, error) {
	// For now, assume cobra commands are not interactive
	// TODO: Add logic to detect interactive commands based on args
	return false, nil
}

func (t *cobraTool) CheckModifiesResource(args map[string]any) string {
	argsVal, ok := args["args"]
	if !ok || argsVal == nil {
		return "unknown"
	}

	argsSlice, ok := argsVal.([]any)
	if !ok || len(argsSlice) == 0 {
		return "unknown"
	}

	// Check the first argument (subcommand) for modification verbs
	if len(argsSlice) > 0 {
		subcommand := strings.ToLower(fmt.Sprintf("%v", argsSlice[0]))
		switch subcommand {
		case "add", "create", "apply", "update", "set", "patch", "edit", "delete", "remove", "rm":
			return "yes"
		case "get", "list", "show", "describe", "status", "info", "view":
			return "no"
		default:
			return "unknown"
		}
	}

	return "unknown"
}
