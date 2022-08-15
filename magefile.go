//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
)

const (
	// collectionName is the name of the Ansible collection.
	collectionName = "delinea.core"
)

func Init() error {
	if err := sh.Run("python3", "-m", "pip", "install", "-r", "requirements.txt"); err != nil {
		pterm.Error.Printfln("error installing requirements: %s", err)
		return err
	}
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

// ➕ InstallBase will install the base Ansible installation based on the provided target version.
func (Ansible) InstallBase(target string) error {

}
