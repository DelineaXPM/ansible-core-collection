//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"

	// mage:import
	"github.com/sheldonhull/magetools/gotools"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

const (
	// collectionName is the name of the Ansible collection.
	collectionName = "delinea.core"

	// VenvDirectory is the directory to keep the Ansible virtual environments.
	VenvDirectory = ".cache"

	// Namespace is the ansible collection namespace.
	Namespace = "delinea"

	// Collection is the ansible collection name.
	Collection = "core"
	// VenvToolingDirectory is the venv for tooling.
	VenvToolingDirectory = "tooling"

	// PermissionUserReadWriteExecute is the octal permission for read, write, & execute only for owner.
	PermissionUserReadWriteExecute = 0o0700

	// PermissionReadWriteSearchAll is the octal permission for all users to read, write, and search a file.
	PermissionReadWriteSearchAll = 0o0777

	// changelogFragments is the directory to store user created changelog fragments.
	changelogFragments = "changelogs/fragments"
)

// AnsibleVersions is a list of Ansible versions to test and create virtual environments for.
var AnsibleVersions = []string{
	"stable-2.10",
	"stable-2.11",
	"stable-2.12",
	"stable-2.13",
	"devel",
}

// Ansible contains the commands for automation with Ansible.
type Ansible mg.Namespace

// Venv contains commands that are specifically isolated to the target venv.
type Venv mg.Namespace

// Py contains the python related commands not specific for venv.
type Py mg.Namespace

// Job contains grouped sets of commands to simplify workflow
type Job mg.Namespace

func checklinux() {
	if runtime.GOOS == "windows" {
		_ = mg.Fatalf(1, "this command is only supported on Linux or Darwin and you are on: %s", runtime.GOOS)
	}
}

func Init() error {
	magetoolsutils.CheckPtermDebug()

	mg.Deps(
		gotools.Go{}.Init,
	)

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

// initVenvParentDirectory is the directory containing all the venv directories for various versions.
func initVenvParentDirectory() error {
	if err := os.MkdirAll(VenvDirectory, 0755); err != nil {
		return err
	}
	return nil
}

// Changelog is the directory for tooling like ansibull-changelog.
//
// The first time you run this, you have to initialize the changelog via: `.cache/tooling/bin/antsibull-changelog init .`
func (Ansible) Changelog() error {
	magetoolsutils.CheckPtermDebug()

	checklinux()
	pterm.DefaultHeader.Println("Changelog()")

	// No fancy semver matching, just trust the input.
	pterm.Info.Println("Enter semver version (example: 1.0.x)")
	versionSemver, _ := pterm.DefaultInteractiveTextInput.
		WithMultiLine(false).Show()
	pterm.Info.Printfln("You answered: %s", versionSemver)

	changelogFragmentFile := filepath.Join(changelogFragments, versionSemver+".yml")

	pterm.Info.Println("Enter release summary")
	releaseNotes, _ := pterm.DefaultInteractiveTextInput.
		WithMultiLine(true).Show()
	pterm.Info.Printfln("You answered: %s", releaseNotes)

	if err := os.WriteFile(changelogFragmentFile, []byte("---\nrelease_summary:\n    "+releaseNotes), PermissionReadWriteSearchAll); err != nil {
		return err
	}
	venvPath := filepath.Join(VenvDirectory, VenvToolingDirectory)
	venvPathBin := filepath.Join(venvPath, "bin")
	pypip := filepath.Join(venvPath, "bin", "pip3")

	if err := sh.Run("python3", "-m", "venv", venvPath); err != nil {
		pterm.Error.Printfln("error installing requirements: %s", err)
		return err
	}
	pterm.Success.Printfln("initialized venvpath: %s", venvPath)
	if err := sh.Run(pypip, "install", "antsibull-changelog", "--disable-pip-version-check"); err != nil {
		pterm.Error.Printfln("error installing wheel in venv %s: %v", venvPath, err)
		return err
	}
	pterm.Success.Println("installed antsibull-changelog")
	// venvPathBin := filepath.Join(venvPath, "bin")
	antsibullchangelog := filepath.Join(venvPath, "bin", "antsibull-changelog")

	pathVar := os.Getenv("PATH")
	newPath := venvPathBin + ":" + pathVar // NOTE: works for mac/linux

	cmd := exec.Cmd{
		Path: antsibullchangelog,
		Args: []string{
			"", // without blank go trims out the subcommand as no flag.
			"release",
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env: []string{
			fmt.Sprintf("PATH=%s", newPath),
			fmt.Sprintf("VIRTUAL_ENV=%s", venvPath),
		},
	}
	pterm.Debug.Printfln("cmd: %v", cmd.String())
	if err := cmd.Run(); err != nil {
		pterm.Warning.Printfln("error: %v", err)
	}
	pterm.Success.Printfln("created venv for: %s", VenvToolingDirectory)
	return nil
}

// 🐍 Init sets up the venv environment (without Ansible yet).
func (Py) Init() error {
	if err := initVenvParentDirectory(); err != nil {
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
	return sh.Run(
		"ansible-test",
		"coverage",
		"xml",
		"-v",
		"--requirements",
		"--group-by",
		"command",
		"--group-by",
		"version",
	)
}

// Setup creates the python venv and installs all the target ansible versions in each.
func (Job) Setup() {
	mg.SerialDeps(
		Py{}.Init,
		Venv{}.Install,
	)
}

// 🧪 TestSanity will run ansible-test with the docker option, using the provided venv.
func (Venv) TestSanity() error {
	magetoolsutils.CheckPtermDebug()
	// needs linux as i don't handle different env path setup
	checklinux()
	prog, _ := pterm.DefaultProgressbar.
		WithTitle("running ansible-test").
		WithTotal(len(AnsibleVersions)).
		WithCurrent(0).
		WithMaxWidth(pterm.GetTerminalWidth() / 2).
		WithTitle("TestSanity").
		WithRemoveWhenDone(false).
		WithShowElapsedTime(true).Start()

	for _, version := range AnsibleVersions {
		venvPath := filepath.Join(VenvDirectory, version)
		venvPathBin := filepath.Join(venvPath, "bin")
		ansibleTest := filepath.Join(venvPath, "bin", "ansible-test")
		activate := filepath.Join(venvPath, "bin", "activate")

		ansibleTestPath, err := filepath.Abs(ansibleTest)
		if err != nil {
			pterm.Warning.Printfln("error in resolving abs ansibleTestPath: %v", err)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		_ = os.Setenv("VIRTUAL_ENV", venvPath)
		pathVar := os.Getenv("PATH")
		newPath := venvPathBin + ":" + pathVar // NOTE: works for mac/linux
		if err := os.Setenv("PATH", newPath); err != nil {
			return err
		}

		pterm.Debug.Printfln("PATH: %s", newPath)
		pterm.Debug.Printfln("running: %s", activate)
		pterm.Debug.Printfln("ansibleTestPath: %s", ansibleTestPath)
		collectionDirectory := filepath.Join(
			homeDir,
			".ansible",
			"collections",
			"ansible_collections",
			Namespace,
			Collection,
		)

		if _, err := os.Stat(collectionDirectory); os.IsNotExist(err) {
			pterm.Error.Println(
				"the target collection doesn't exist. It's likey you need to run:\n\n\tmage ansible:installcollection",
			)
		}
		prog.UpdateTitle(fmt.Sprintf("ansible-test: %s", version))

		cmd := exec.Cmd{
			Path: ansibleTestPath,
			Dir:  collectionDirectory,
			Args: []string{
				"",
				"sanity",
				"--docker",
				"-v",
				"--color",
				"--coverage",
			}, // empty string required to avoid subcommand without flags disappearing
			Stdout: nil,
			Stderr: os.Stderr,
			Env: []string{
				fmt.Sprintf("PATH=%s", newPath),
				fmt.Sprintf("VIRTUAL_ENV=%s", venvPath),
				fmt.Sprintf("HOME=%s", homeDir),
			},
		}
		pterm.Debug.Printfln("cmd: %v", cmd.String())

		prog.Increment()

		if err := cmd.Run(); err != nil {
			pterm.Warning.Printfln("error: %v", err)
		}

	}
	return nil
}
