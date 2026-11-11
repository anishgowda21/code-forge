package git

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

func run(ctx context.Context, cmdArgs []string) (string, error) {
	if len(cmdArgs) < 2 {
		return "", errors.New("invalid command")
	}
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)

	cmd.WaitDelay = 5 * time.Second

	output, err := cmd.CombinedOutput()
	return string(output), err
}
