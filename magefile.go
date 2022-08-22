//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
)

const (
	// collectionName is the name of the Ansible collection.
	collectionName = "delinea.core"

	// VenvDirectory is the directory to keep the Ansible virtual environments.
	VenvDirectory = ".cache"
)

var (

	// AnsibleVersions is a list of Ansible versions to test and create virtual environments for.
	AnsibleVersions = []string{
		"stable-2.10",
		"stable-2.11",
		"stable-2.12",
		"devel",
	}
)

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

// Ansible contains the commands for automation with Ansible.
type Ansible mg.Namespace

// ➕ InstallCollection will install the collection.
func (Ansible) InstallCollection() error {
	return sh.Run("ansible-galaxy", "collection", "install", collectionName)
}

// ➕ InstallCollection will install the collection.
func (Ansible) UninstallCollection() error {
	return sh.Run("ansible-galaxy", "collection", "install", collectionName)
}

type Python3 mg.Namespace

// 🐍 Init sets up the venv environment (without Ansible yet).
func (Python3) Init() error {
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
	pterm.Success.Println("(Python3) Init()")
	return nil
}

func (Ansible) InstallInVenv() error {
	if err := os.MkdirAll(VenvDirectory, 0755); err != nil {
		return err
	}

	downloadLink := "https://github.com/ansible/ansible/archive/%s.tar.gz"

	for _, version := range AnsibleVersions {
		venvPath := filepath.Join(VenvDirectory, version)
		pypip := filepath.Join(venvPath, "bin", "pip")

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
	pterm.Success.Println("(Python3) Init()")
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
	return sh.Run("python3", "-m", "ansible-test", "sanity", "--docker", "-v", "--color", "--coverage")
}

// 🧪 TestUnit will run ansible-test with the docker option.
func (Ansible) TestUnit() error {
	return sh.Run("python3", "-m", "ansible-test", "unit", "--docker", "-v", "--color", "--coverage")
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
	return sh.Run("python3", "-m", "ansible-test", "coverage", "coverage", "xml", "-v", "--requirements", "--group-by", "command", "--group-by", "version")
}
