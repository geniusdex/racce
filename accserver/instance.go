package accserver

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstanceState is the state of a possible running instance
type InstanceState int

const (
	Stopped InstanceState = iota
	Stopping
	Running
)

// Instance represents a running instance of an accServer
type Instance struct {
	cmd       *exec.Cmd
	hasKilled bool
}

func makeCmd(accServer string) *exec.Cmd {
	cmd := exec.Command(accServer)
	if wine, err := exec.LookPath("wine"); err == nil {
		cmd = exec.Command(wine, accServer)
	}
	cmd.Dir = filepath.Dir(accServer)
	return cmd
}

func cmdString(cmd *exec.Cmd) string {
	// cmd.Path is the first element in cmd.Args
	return "'" + strings.Join(cmd.Args, "' '") + "'"
}

func newInstance(executable string) (*Instance, error) {
	i := &Instance{
		cmd:       makeCmd(executable),
		hasKilled: false,
	}

	log.Printf("Starting %s...", cmdString(i.cmd))

	if err := i.cmd.Start(); err != nil {
		return nil, err
	}

	go i.wait()

	return i, nil
}

// State resolves the current state of an instance. A nil instance is accepted and
// resolves to Stopped.
func (i *Instance) State() InstanceState {
	if i == nil || i.cmd.Process == nil || i.cmd.ProcessState != nil {
		return Stopped
	}
	if i.hasKilled {
		return Stopping
	}
	return Running
}

// IsRunning returns if the instance is running (State() == Running)
func (i *Instance) IsRunning() bool {
	return i.State() == Running
}

// IsStopped returns if the instance is stopped (State() == Stopped)
func (i *Instance) IsStopped() bool {
	return i.State() == Stopped
}

// IsStopping returns if the instance is stopping (State() == Stopping)
func (i *Instance) IsStopping() bool {
	return i.State() == Stopping
}

// stop kills the running instance
func (i *Instance) stop() error {
	if i.State() == Stopped {
		return fmt.Errorf("server is already stopped")
	}
	log.Printf("Killing accServer...")
	i.hasKilled = true
	return i.cmd.Process.Kill()
}

// wait for the process to terminate and then update the state accordingly
func (i *Instance) wait() {
	if err := i.cmd.Wait(); err != nil {
		log.Printf("Error waiting for accServer process to exit: %v", err)
	} else {
		log.Printf("The accServer process has exited normally")
	}
}
