package portforward

import (
	"context"
	"fmt"
	"os"
	"os/exec"
  "sync"
  "syscall"
  // "path/filepath"
)
var mu sync.Mutex
func PortForward(ctx context.Context, localIP, localPort, remoteIP, remotePort string) (int, error) {
  mu.Lock()
  defer mu.Unlock()
  fmt.Println("im herer")
  // binaryPath := "/home/panipuri/Desktop/workspace/backend/portforward/tcp"
	args := []string{
		fmt.Sprintf("--bind-address=%s", localIP),
		fmt.Sprintf("--local-port=%s", localPort),
		fmt.Sprintf("--remote-host=%s", remoteIP),
		fmt.Sprintf("--remote-port=%s", remotePort),
	}
  fmt.Println("im here:", "./tcp", args)
  // dir, err := os.Getwd()
	// if err != nil {
	// 	return -1, fmt.Errorf("failed to get current directory: %v", err)
	// }
  // tcpPath := filepath.Join(dir, "tcp")
	cmd := exec.CommandContext(ctx, "./tcp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
 cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} //
	err := cmd.Start()
	if err != nil {
    fmt.Printf("[ERROR] Failed to start ./main: %v\n", err)
		return 0, err
	}
  go func() {
      _ = cmd.Wait() // prevents zombie ps
  }()
  fmt.Printf("[INFO] Started port forward with PID %d\n", cmd.Process.Pid)
	return cmd.Process.Pid, nil
}
