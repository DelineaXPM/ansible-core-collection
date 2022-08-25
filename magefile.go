//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"

	"github.com/bitfield/script"
)

const (
	// collectionName is the name of the Ansible collection.
	collectionName = "delinea.core"

	// VenvDirectory is the directory to keep the Ansible virtual environments.
	VenvDirectory = ".cache"

	// Namespace is the ansible collection namespace.
	Namespace = "delinea"

	//Collection is the ansible collection name.
	Collection = "core"
)

var (

	// AnsibleVersions is a list of Ansible versions to test and create virtual environments for.
	AnsibleVersions = []string{
		"stable-2.10",
		"stable-2.11",
		"stable-2.12",
		"stable-2.13",
		"devel",
	}
)

// Ansible contains the commands for automation with Ansible.
type Ansible mg.Namespace

// Venv contains commands that are specifically isolated to the target venv.
type Venv mg.Namespace

// Py contains the python related commands not specific for venv.
type Py mg.Namespace

// Job contains grouped sets of commands to simplify workflow
type Job mg.Namespace

func Init() error {

	pterm.Success.Println("Init()")
	return nil
}

// Clean removes the local .artifact and .cache/ directories.
func Clean() {
	_ = os.RemoveAll(".artifacts/")
	_ = os.RemoveAll(".cache/")
	os.Mkdir(".artifacts/", 0755)
	os.Mkdir(".cache/", 0755)
	pterm.Success.Println("reset .artifacts and .cache/ directories")
}

// ➕ InstallCollection will install the collection.
func (Ansible) InstallCollection() error {
	return sh.Run("ansible-galaxy", "collection", "install", collectionName)
}

// ➕ InstallCollection will install the collection.
func (Ansible) UninstallCollection() error {
	return sh.Run("ansible-galaxy", "collection", "install", collectionName)
}

// 🐍 Init sets up the venv environment (without Ansible yet).
func (Py) Init() error {
	if err := os.MkdirAll(VenvDirectory, 0755); err != nil {
		return err
	}
	for _, version := range AnsibleVersions {
		if err := sh.Run("python3", "-m", "venv", filepath.Join(VenvDirectory, version)); err != nil {
			pterm.Error.Printfln("error installing requirements: %s", err)
			return err
		}
		pterm.Success.Printfln("created venv for: %s", version)
	}
	pterm.Success.Println("(Py) Init()")
	return nil
}

func (Venv) Install() error {
	if err := os.MkdirAll(VenvDirectory, 0755); err != nil {
		return err
	}

	downloadLink := "https://github.com/ansible/ansible/archive/%s.tar.gz"

	for _, version := range AnsibleVersions {
		venvPath := filepath.Join(VenvDirectory, version)
		pypip := filepath.Join(venvPath, "bin", "pip3")

		pterm.Info.Printfln("installing requirements in venv: %s", venvPath)

		err := sh.Run(pypip, "install", "wheel", "--disable-pip-version-check")
		if err != nil {
			pterm.Error.Printfln("error installing wheel in venv %s: %v", venvPath, err)
		}

		err = sh.Run(pypip, "install", fmt.Sprintf(downloadLink, version), "--disable-pip-version-check")
		if err != nil {
			pterm.Error.Printfln("error installing ansible in venv %s: %v", venvPath, err)
		}

		pterm.Success.Printfln("created venv for: %s", version)
	}
	pterm.Success.Println("(Venv) Init()")
	return nil
}

// ➕ InstallBase (parameters: target) will install the base Ansible installation based on the provided target version.
func (Ansible) InstallBase(target string) error {
	if target == "" {
		pterm.Error.Println("no target was provided, can't proceed")
		pterm.Error.Println("You might try providing a value such as: \n\n" +
			"- stable-2.10\n" +
			"- stable-2.11\n" +
			"- stable-2.12\n" +
			"- stable-2.13\n" +
			"- devel\n",
		)
		return fmt.Errorf("missing parameter for InstallBase")
	}
	return sh.RunV(
		"python3", "-m", "pip",
		"install", fmt.Sprintf("https://github.com/ansible/ansible/archive/%s.tar.gz", target),
		"--disable-pip-version-check",
		"--user",
	)
}

// 🧪 TestSanity will run ansible-test with the docker option.
func (Ansible) TestSanity() error {
	return sh.Run("ansible-test", "sanity", "--docker", "-v", "--color", "--coverage")
}

// 🧪 TestUnit will run ansible-test with the docker option.
func (Ansible) TestUnit() error {
	return sh.Run("ansible-test", "unit", "--docker", "-v", "--color", "--coverage")
}

// 🧪 Test will run both unit and Sanity tests.
func (Ansible) Test() {
	mg.SerialDeps(
		Ansible.TestSanity,
		Ansible.TestUnit,
	)
}

// 📈 Coverage will run generate code coverage data for ansible-test.
func (Ansible) Coverage() error {
	return sh.Run("ansible-test", "coverage", "coverage", "xml", "-v", "--requirements", "--group-by", "command", "--group-by", "version")
}

// Setup creates the python venv and installs all the target ansible versions in each.
func (Job) Setup() {

	mg.SerialDeps(
		Py{}.Init,
		Venv{}.Install,
	)
}

// 🧪 TestSanity will run ansible-test with the docker option.
func (Venv) TestSanity() error {

	for _, version := range AnsibleVersions {
		venvPath := filepath.Join(VenvDirectory, version)
		ansibleTest := filepath.Join(venvPath, "bin", "ansible-test")
		activate := filepath.Join(venvPath, "bin", "activate")

		// deactivate := filepath.Join(venvPath, "bin", "deactivate")

		// if err := sh.Run(activate); err != nil {
		// 	return err
		// }
		ansibleTestPath, err := filepath.Abs(ansibleTest)
		if err != nil {
			pterm.Warning.Printfln("error in creating ansibleTestPath: %v", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		// cmd := fmt.Sprintf("%s sanity --docker -v --color --coverage", ansibleTestPath)
		// pterm.Info.Printfln("running: %s", cmd)
		script.Exec(activate).Stdout()
		pterm.Info.Printfln("ansibletestPath: %s", ansibleTestPath)
		cmd := exec.Cmd{
			Path:   ansibleTestPath,
			Dir:    filepath.Join(homeDir, ".ansible", "collections", "ansible_collections", Namespace, Collection),
			Args:   []string{"", "sanity", "--docker", "-v", "--color", "--coverage"},
			Stdout: os.Stdout,
			Stderr: os.Stdout,
		}

		pterm.Info.Printfln("cmd: %v", cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
		// if err := sh.Run(ansibleTest, "sanity", "--docker", "-v", "--color", "--coverage"); err != nil {
		// 	return err
		// }
		// if err := sh.Run(deactivate); err != nil {
		// 	return err
		// }
	}
	return nil
}
