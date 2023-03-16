//go:build mage

package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	// "github.com/bitfield/script"
	"github.com/bitfield/script"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"

	"github.com/Masterminds/semver/v3"
	"github.com/sheldonhull/magetools/ci"
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

	// PermissionUserReadWriteExecuteGroupReadOnly Chmod 0755 (chmod a+rwx,g-w,o-w,ug-s,-t) sets permissions so that, (U)ser / owner can read, can write and can execute. (G)roup can read, can't write and can execute. (O)thers can read, can't write and can execute.
	PermissionUserReadWriteExecuteGroupReadOnly = 0o755

	// PermissionReadWriteSearchAll is the octal permission for all users to read, write, and search a file.
	PermissionReadWriteSearchAll = 0o0777

	// changelogFragments is the directory to store user created changelog fragments.
	changelogFragments = "changelogs/fragments"

	// ArtifactDirectory is the directory to store artifacts and ignored by git.
	ArtifactDirectory = ".artifacts"

	// GalaxyYaml is the name of the galaxy.yml file.
	GalaxyYaml = "galaxy.yml"
)

// AnsibleVersions is a list of Ansible versions to test and create virtual environments for.
var AnsibleVersions = []string{
	"stable-2.10",
	"stable-2.11",
	"stable-2.12",
	"stable-2.13",
	"devel",
}

// AnsibleVersionCI is the version of Ansible to use for CI.
var AnsibleVersionCI = "stable-2.13"

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

// Release runs all the required steps to create setup the virtual env for publishing and create the release artifact.
func (Job) Release() {
	magetoolsutils.CheckPtermDebug()
	checklinux()

	mg.SerialDeps(
		Py.InitSingle,
		Venv.InstallSingle,
		// Ansible.Doctor
		Release,
	)
}

func Init() {
	magetoolsutils.CheckPtermDebug()
	createDirectories()

	mg.Deps(
		gotools.Go{}.Init,
	)

	if ci.IsCI() {
		pterm.Info.Println("CI detected, skipping remaining steps for Init()")
		return
	}

	// setup venv and requirements

	mg.Deps(Job{}.Setup)
	pterm.Info.Printfln("if you want to activate an environment manually, run on of the source commands printed to activate in your terminal")

	pterm.Success.Println("Init()")
}

// DeepClean removes not only artifacts, but the cached python virtual environments.
// Since this can take 8 mins to resetup, it's not in the default clean actions.
func DeepClean() {
	_ = os.RemoveAll(ArtifactDirectory)
	_ = os.RemoveAll(VenvDirectory)
	createDirectories()
	pterm.Success.Println("DeepClean() reset .artifacts, .cache/, and venv")
}

// Clean removes the local .artifact and .cache/ directories.
func Clean() {
	_ = os.RemoveAll(ArtifactDirectory)
	createDirectories()
	pterm.Success.Println("Clean() reset .artifacts and .cache/")
}

func createDirectories() {
	if err := os.Mkdir(ArtifactDirectory, PermissionUserReadWriteExecuteGroupReadOnly); err != nil && !errors.Is(err, os.ErrExist) {
		pterm.Error.Printfln("failed to create .artifacts/ directory: %s", err)
	}
	if err := os.Mkdir(VenvDirectory, PermissionUserReadWriteExecuteGroupReadOnly); err != nil && !errors.Is(err, os.ErrExist) {
		pterm.Error.Printfln("failed to create .cache/ directory: %s", err)
	}
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
	if err := os.MkdirAll(VenvDirectory, PermissionUserReadWriteExecuteGroupReadOnly); err != nil {
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

// 🐍 InitSingle sets up a single python virtual environment for publishing or other actions.
func (Py) InitSingle() error {
	pterm.DefaultHeader.Println("InitSingle()")
	if err := initVenvParentDirectory(); err != nil {
		return err
	}

	if err := sh.Run("python3", "-m", "venv", filepath.Join(VenvDirectory, AnsibleVersionCI)); err != nil {
		pterm.Error.Printfln("error installing requirements: %s", err)
		return err
	}
	pterm.Success.Printfln("created venv for: %s", AnsibleVersionCI)

	pterm.Success.Println("(Py) InitSingle()")
	return nil
}

func (Venv) Install() error {
	if err := os.MkdirAll(VenvDirectory, PermissionUserReadWriteExecuteGroupReadOnly); err != nil {
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
		pterm.Info.Printfln("source .cache/%s/bin/activate", version)
	}

	pterm.Success.Println("(Venv) Init()")
	return nil
}

// InstallSingle sets up a single python virtual environment for publishing or other actions.
func (Venv) InstallSingle() error {
	pterm.DefaultHeader.Println("InstallSingle()")
	if err := os.MkdirAll(VenvDirectory, PermissionUserReadWriteExecuteGroupReadOnly); err != nil {
		return err
	}

	downloadLink := "https://github.com/ansible/ansible/archive/%s.tar.gz"
	venvPath := filepath.Join(VenvDirectory, AnsibleVersionCI)
	pypip := filepath.Join(venvPath, "bin", "pip3")
	pterm.Info.Printfln("installing requirements in venv: %s", venvPath)
	err := sh.Run(pypip, "install", "wheel", "--disable-pip-version-check")
	if err != nil {
		pterm.Error.Printfln("error installing wheel in venv %s: %v", venvPath, err)
	}

	err = sh.Run(pypip, "install", fmt.Sprintf(downloadLink, AnsibleVersionCI), "--disable-pip-version-check")
	if err != nil {
		pterm.Error.Printfln("error installing ansible in venv %s: %v", venvPath, err)
	}

	pterm.Success.Printfln("created venv for: %s", AnsibleVersionCI)
	pterm.Info.Printfln("source .cache/%s/bin/activate", AnsibleVersionCI)

	pterm.Success.Println("(Venv) InitSingle()")
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
			"- devel",
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
	pterm.Warning.Printfln("this might take up to 8 minutes as it sets up virtual environments for multiple versions")
	createDirectories()

	mg.SerialDeps(
		Py{}.Init,
		Venv{}.Install,
	)
}

// 🧪 TestSanity will run ansible-test with the docker option against all available versions.
func (Venv) TestSanity() error {
	magetoolsutils.CheckPtermDebug()
	// needs linux as i don't handle different env path setup
	checklinux()
	prog, _ := pterm.DefaultProgressbar.
		WithTitle("running ansible-test").
		WithTotal(len(AnsibleVersions)).
		WithCurrent(0).
		WithMaxWidth(pterm.GetTerminalWidth() / 2). //nolint:gomnd // allow magic num
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
		pterm.Debug.Printfln("collectionDirectory: %q", collectionDirectory)
		if _, err := os.Stat(collectionDirectory); os.IsNotExist(err) {
			pterm.Error.Println(
				"the target collection doesn't exist. It's likey you need to run:\n\n\tmage ansible:installcollection",
			)
		}
		prog.UpdateTitle(fmt.Sprintf("ansible-test: %s", version))
		pterm.Debug.Printfln("To run a local test outside mage change directories to collectionDirectory, and then run the command debug will output")
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
				"--skip-test",
				"symlinks,shebang", // causes issues with project files like devcontainer
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

// 🚀 Release will publish the release to Ansible Galaxy.
func Release() error {
	pterm.DefaultHeader.Println("Release()")
	venvPath := filepath.Join(VenvDirectory, AnsibleVersionCI)
	galaxyCLI := filepath.Join(venvPath, "bin", "ansible-galaxy")

	if err := sh.RunV(galaxyCLI, "--version"); err != nil {
		return err
	}
	pterm.Success.Println("Printed version")

	pterm.DefaultSection.Println("Building...")
	if err := sh.RunV(galaxyCLI,
		"collection",
		"build",
		"-v",
		"--force",
		"--output-path", filepath.Join(ArtifactDirectory, "")); err != nil {
		return err
	}
	pterm.Success.Println("Built collection")

	archivePattern := filepath.Join(ArtifactDirectory, "delinea-core*.tar.gz")
	archive, err := filepath.Glob(archivePattern)
	if err != nil {
		return err
	}

	if len(archive) == 0 {
		return fmt.Errorf("no archive found with pattern %q", archivePattern)
	}
	pterm.Success.Printfln("archive found: %v", archive)
	archiveName := archive[0]
	pterm.DefaultSection.Printf("Archive name: %s\n", archiveName)

	r, err := os.Open(archiveName)
	if err != nil {
		return err
	}
	archivereader, err := gzip.NewReader(r)

	reader := tar.NewReader(archivereader)
	pterm.DefaultSection.Println("List Archive content")
	for {
		header, err := reader.Next()
		// use errors.Is to check for EOF, since raw check of errors won't handle wrapped errors
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(header.Name)
	}
	pterm.Success.Println("Viewed archive content")

	pterm.DefaultSection.Println("Publishing...")
	if os.Getenv("GALAXY_KEY") == "" || os.Getenv("GALAXY_SERVER") == "" {
		pterm.Warning.Printfln("GALAXY_KEY or GALAXY_SERVER not set. Skipping publish")
		return fmt.Errorf("missing required environment variables, did you run ansible:doctor first?")
	}
	if err := sh.RunV(galaxyCLI, "collection", "publish", "-v", "--server", os.Getenv("GALAXY_SERVER"), "--api-key", os.Getenv("GALAXY_KEY"), archiveName); err != nil {
		return err
	}
	pterm.Success.Println("Published collection")

	return nil
}

// checkEnv is the struct to pass into the checkEnvVar function to check and validate the environment variables.
// This builds a nice table summary when used to help summarize all the failed checks rather than doing this piecemeal.
type checkEnv struct {
	Name       string
	IsSecret   bool
	IsRequired bool
	Tbl        pterm.TableData
	Notes      string
}

// checkEnvVar performs a check on environment variable and helps build a report summary of the failing conditions, missing variables, and bypasses logging if it's a secret.
// Yes this could be replaced by the `env` package but I had this in place and the output is nice for debugging so I left it. - Sheldon 😀
//
//nolint:unparam // ignoring as i'll want to use the values in the future, ok to leave for now.
func checkEnvVar(ckv checkEnv) (value string, ptd pterm.TableData, err error) {
	// loggedValue is used to make sure any secret isn't put into the table output.
	var loggedValue string
	var isSet bool
	tbl := ckv.Tbl
	value, isSet = os.LookupEnv(ckv.Name)
	if isSet {
		if ckv.IsSecret {
			loggedValue = "***** secret set, but not logged *****"
		} else {
			loggedValue = value
		}
	}
	// Required but not set is an error condition to report back to the user.
	if !isSet && ckv.IsRequired {
		tbl = append(tbl, []string{"❌", ckv.Name, loggedValue, ckv.Notes})
		return "", tbl, fmt.Errorf("%s is required and not set", ckv.Name)
	}
	// Required but not a terminating error, then just put as information different from success, and no error.
	if !isSet && !ckv.IsRequired {
		tbl = append(tbl, []string{"👉", ckv.Name, loggedValue, ckv.Notes})
		return value, tbl, nil
	}

	if isSet {
		tbl = append(tbl, []string{"✅", ckv.Name, loggedValue, ckv.Notes})
		return value, tbl, nil
	}
	return "", tbl, fmt.Errorf("unknown error (no conditions were hit so it's a PEKAB issue 😁) with evaluation of: %s", ckv.Name)
}

// Doctor will validate the required tools and environment variables are available.
func (Ansible) Doctor() error {
	pterm.DefaultHeader.Printfln("CheckEnvironment")
	tbl := pterm.TableData{
		[]string{"Status", "Check", "Value", "Notes"},
		[]string{"✅", "GOOS", runtime.GOOS, ""},
		[]string{"✅", "GOARCH", runtime.GOARCH, ""},
		[]string{"✅", "GOROOT", runtime.GOROOT(), ""},
		[]string{"✅", "GOPATH", os.Getenv("GOPATH"), ""},
	}

	defer func(tbl *pterm.TableData) {
		primary := pterm.NewStyle(pterm.FgLightWhite, pterm.BgGray, pterm.Bold)

		if err := pterm.DefaultTable.WithHasHeader().
			WithBoxed(true).
			WithHeaderStyle(primary).
			WithData(*tbl).Render(); err != nil {
			pterm.Error.Printf(
				"pterm.DefaultTable.WithHasHeader of variable information failed. Continuing...\n%v",
				err,
			)
		}
	}(&tbl)
	var errorCount int
	_, tbl, err := checkEnvVar(checkEnv{Name: "GALAXY_SERVER", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required for defining target publish location"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(checkEnv{Name: "GALAXY_KEY", IsSecret: true, IsRequired: true, Tbl: tbl, Notes: "required for publishing"})
	if err != nil {
		errorCount++
	}

	output, err := sh.Output("ansible-galaxy", "--version")
	if err != nil {
		errorCount++
		tbl = append(tbl, []string{"❌", "ansible-galaxy", "ansible-galaxy", "required for building and publishing"})
	} else {
		re := regexp.MustCompile(`(?m)^ansible-galaxy.*$`)
		match := re.FindString(output)
		pterm.Debug.Printfln("output: %s", output)

		tbl = append(tbl, []string{"✅", "ansible-galaxy", match, "required cli tool"})
	}
	// using sh.Output to get the version of python3, and append to tbl
	version, err := sh.Output("python3", "--version")
	if err != nil {
		errorCount++
		tbl = append(tbl, []string{"❌", "python3", "python3", "required for building and publishing"})
	} else {
		tbl = append(tbl, []string{"✅", "python3", version, "required for building and publishing"})
	}

	if errorCount > 0 {
		pterm.Error.Printfln("required checks failed: %d", errorCount)
		return fmt.Errorf("failed %d checks", errorCount)
	}
	return nil
}

// Bump will bump the version of the collection, using yq and updating the galaxy version.
// Valid types are "major", "minor", "patch"
func (Ansible) Bump(bumpType string) error {
	pterm.DefaultHeader.Printfln("BumpVersion")
	pterm.Info.Printfln("bumpType: %s", bumpType)
	// read the current version
	currentVersion, err := script.Exec(fmt.Sprintf("yq \".version\" %s", GalaxyYaml)).String()
	if err != nil {
		pterm.Error.Printfln("failed to get version from galaxy.yml: %v", err)
		return err
	}
	pterm.Info.Printfln("current version: %s", currentVersion)
	// Parse the current version number from the structured object
	version, err := semver.StrictNewVersion(strings.TrimSpace(currentVersion))
	if err != nil {
		return err
	}
	pterm.Info.Printfln("parsed version: %s", version.String())

	var newVersion semver.Version
	// use semver type to correctly increment based on desired type.
	switch bumpType {
	case "major":
		// Increment the major version number
		newVersion = version.IncMajor()
	case "minor":
		// Increment the minor version number
		newVersion = version.IncMinor()
	case "patch":
		// Increment the patch version number
		newVersion = version.IncPatch()
	default:
		return fmt.Errorf("unknown bump type: %s", bumpType)
	}

	// update the current version and replace yaml
	pterm.Info.Printfln("new version: %s", newVersion.String())
	commandToRun := fmt.Sprintf("yq --inplace '.version = \"%s\" ' %s", newVersion.String(), GalaxyYaml)
	pterm.Info.Printfln("command: %s", commandToRun)
	_, err = script.Exec(commandToRun).Stdout()
	if err != nil {
		pterm.Error.Printfln("failed to get version from galaxy.yml: %v", err)
		return err
	}
	return nil
}
