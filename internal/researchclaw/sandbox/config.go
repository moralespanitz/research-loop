// Package sandbox provides configuration and lifecycle management for
// isolated experiment execution environments within Research Loop.
//
// Four sandbox modes are supported:
//   - Local: executes experiments as a subprocess in a Python venv
//   - Docker: runs experiments in an OCI container with resource limits
//   - SSH Remote: dispatches experiments to a remote server via SSH
//   - Colab: bridges experiments to Google Colab via Google Drive
//
// Each mode isolates experiments from the host system to varying degrees.
// Docker offers the strongest isolation and reproducibility guarantee,
// while local mode is fastest for iterative development.
//
// File layout (future):
//   config.go    — SandboxConfig struct and validation
//   runner.go    — ExperimentRunner interface and implementations per mode
//   validator.go — Code validation (syntax, security, import checks)
package sandbox

// Mode enumerates the supported sandbox execution modes.
type Mode string

const (
	ModeLocal      Mode = "local"       // Python venv subprocess
	ModeDocker     Mode = "docker"      // OCI container with resource limits
	ModeSSHRemote  Mode = "ssh_remote"  // Remote server via SSH
	ModeColab      Mode = "colab"       // Google Colab via Drive bridge
)

// SandboxConfig holds all configuration for an experiment sandbox instance.
// The Mode field selects which execution backend to use. Other fields are
// mode-specific and may be empty for unrelated modes.
type SandboxConfig struct {
	// Mode selects the execution backend: local, docker, ssh_remote, or colab.
	Mode Mode `mapstructure:"mode" json:"mode" yaml:"mode"`

	// PythonPath is the path to the Python interpreter for local mode.
	// Default: ".venv/bin/python3" (Unix) or ".venv/Scripts/python.exe" (Windows).
	PythonPath string `mapstructure:"python_path" json:"python_path" yaml:"python_path"`

	// TimeoutSec is the maximum wall-clock time (in seconds) for a single
	// experiment run. The experiment's time guard should fire at 80% of this.
	TimeoutSec int `mapstructure:"timeout_sec" json:"timeout_sec" yaml:"timeout_sec"`

	// MaxMemoryMB is the maximum memory allocation for the sandbox.
	// In Docker mode this maps to --memory. In local mode it is advisory.
	MaxMemoryMB int `mapstructure:"max_memory_mb" json:"max_memory_mb" yaml:"max_memory_mb"`

	// GPURequired indicates whether GPU devices are needed. In Docker mode
	// this enables --gpus all. In SSH remote mode it sets CUDA_VISIBLE_DEVICES.
	GPURequired bool `mapstructure:"gpu_required" json:"gpu_required" yaml:"gpu_required"`

	// DockerImage is the OCI image name for Docker mode.
	// Example: "research-loop/experiment:latest"
	DockerImage string `mapstructure:"docker_image" json:"docker_image" yaml:"docker_image"`

	// DockerNetworkPolicy controls network access in Docker containers.
	// Values: "none" (no network), "setup_only" (pip install only), "full".
	DockerNetworkPolicy string `mapstructure:"docker_network_policy" json:"docker_network_policy" yaml:"docker_network_policy"`

	// AutoInstallDeps, if true, runs pip install on experiment imports before
	// execution. Only applies in Docker and local modes.
	AutoInstallDeps bool `mapstructure:"auto_install_deps" json:"auto_install_deps" yaml:"auto_install_deps"`

	// KeepContainers, if true, preserves Docker containers after execution
	// for debugging. Only applies in Docker mode.
	KeepContainers bool `mapstructure:"keep_containers" json:"keep_containers" yaml:"keep_containers"`

	// SSHHost is the hostname or IP for SSH remote mode.
	SSHHost string `mapstructure:"ssh_host" json:"ssh_host" yaml:"ssh_host"`

	// SSHUser is the username for SSH remote mode (default: current user).
	SSHUser string `mapstructure:"ssh_user" json:"ssh_user" yaml:"ssh_user"`

	// SSHPort is the SSH port (default: 22).
	SSHPort int `mapstructure:"ssh_port" json:"ssh_port" yaml:"ssh_port"`

	// SSHKeyPath is the path to the SSH private key (default: ~/.ssh/id_rsa).
	SSHKeyPath string `mapstructure:"ssh_key_path" json:"ssh_key_path" yaml:"ssh_key_path"`

	// RemoteWorkdir is the working directory on the remote host.
	RemoteWorkdir string `mapstructure:"remote_workdir" json:"remote_workdir" yaml:"remote_workdir"`

	// RemotePython is the Python interpreter path on the remote host.
	RemotePython string `mapstructure:"remote_python" json:"remote_python" yaml:"remote_python"`

	// UseDockerOverSSH, when true, runs experiments inside Docker on the
	// remote host. Combines Docker isolation with remote GPU access.
	UseDockerOverSSH bool `mapstructure:"use_docker_over_ssh" json:"use_docker_over_ssh" yaml:"use_docker_over_ssh"`

	// ColabDriveRoot is the local path to the Google Drive mount for Colab mode.
	ColabDriveRoot string `mapstructure:"colab_drive_root" json:"colab_drive_root" yaml:"colab_drive_root"`

	// ColabPollIntervalSec is how often to check for experiment results in
	// Google Drive (Colab mode).
	ColabPollIntervalSec int `mapstructure:"colab_poll_interval_sec" json:"colab_poll_interval_sec" yaml:"colab_poll_interval_sec"`

	// ColabTimeoutSec is the maximum wait for a Colab experiment to complete.
	ColabTimeoutSec int `mapstructure:"colab_timeout_sec" json:"colab_timeout_sec" yaml:"colab_timeout_sec"`

	// ColabSetupScript contains shell commands to run before each experiment
	// in Colab mode (e.g., "pip install torch").
	ColabSetupScript string `mapstructure:"colab_setup_script" json:"colab_setup_script" yaml:"colab_setup_script"`
}

// DefaultSandboxConfig returns a SandboxConfig with sensible defaults for
// local development: local mode, 300s timeout, 4GB memory, venv python.
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		Mode:           ModeLocal,
		PythonPath:     ".venv/bin/python3",
		TimeoutSec:     300,
		MaxMemoryMB:    4096,
		GPURequired:    false,
		AutoInstallDeps: true,
		KeepContainers: false,
		SSHPort:        22,
		RemoteWorkdir:  "/tmp/research-loop-experiments",
		RemotePython:   "python3",
		ColabPollIntervalSec: 30,
		ColabTimeoutSec:      3600,
	}
}
