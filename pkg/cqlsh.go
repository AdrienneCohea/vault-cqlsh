package pkg

import (
	"os"
	"os/exec"
	"syscall"
)

func ExecCqlsh(username, password, host, port string, extraArgs []string) error {

	cqlsh, err := exec.LookPath("cqlsh")
	if err != nil {
		return err
	}

	args := []string{"cqlsh", "--username", username, "--password", password}
	for _, arg := range extraArgs {
		args = append(args, arg)
	}
	args = append(args, host)
	if port != "" {
		args = append(args, port)
	}

	env := os.Environ()

	return syscall.Exec(cqlsh, args, env)
}
