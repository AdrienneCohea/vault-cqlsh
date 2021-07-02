package pkg

import (
	"log"
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
	args = append(args, extraArgs...)
	args = append(args, host)
	// if port != "" {
	// 	args = append(args, port)
	// }

	env := os.Environ()

	log.Printf("cqlsh: Starting process with %+v\n", args)

	return syscall.Exec(cqlsh, args, env)
}
