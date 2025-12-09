package servicetools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

var mu sync.Mutex

func PortForward(ctx context.Context, localIP, localPort, remoteIP, remotePort string) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	args := []string{
		fmt.Sprintf("--bind-address=%s", localIP),
		fmt.Sprintf("--local-port=%s", localPort),
		fmt.Sprintf("--remote-host=%s", remoteIP),
		fmt.Sprintf("--remote-port=%s", remotePort),
	}
	cmd := exec.CommandContext(ctx, "./tcp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start ./tcp: %v\n", err)
		return 0, err
	}
	go func() {
		_ = cmd.Wait() // prevents zombie ps
	}()
	fmt.Printf("Started port forward with PID %d\n", cmd.Process.Pid)
	return cmd.Process.Pid, nil
}
