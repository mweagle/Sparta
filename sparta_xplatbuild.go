// +build  !lambdabinary

package sparta

import "syscall"

// Support Windows development, by only requiring `syscall` in the compiled
// linux binary.
func platformKill(parentProcessPID int) {
	syscall.Kill(parentProcessPID, syscall.SIGUSR2)
}
