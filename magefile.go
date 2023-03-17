//go:build mage

package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	//"os/exec"
	//"regexp"
	//"runtime"
	//"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

const (
	// AnsibleLatest defines the latest stable version we use and therefore support.
	AnsibleLatest = "stable-2.13"

	// CacheDir is the directory to keep virtual environments (ignored by git).
	CacheDir = ".cache"

	// ArtifactDir is the directory to store artifacts (ignored by git).
	ArtifactDir = ".artifacts"
)

// ✨ Init unfolds initial environment for productive work.
func Init() error {
	magetoolsutils.CheckPtermDebug()

	if ci.IsCI() {
		pterm.Error.Println("CI should explicitly call `mage initCI <version_name>`")
		return nil
	}

	return ansibleInit(AnsibleLatest)
}

// 🎩 InitCI initializes a new Python virtual environment with given version of Ansible installed.
func InitCI(version string) error {
	return ansibleInit(version)
}

// 🧹 Clean removes '.artifact/', '.cache/', 'tests/output/', directories from the project.
func Clean() {
	magetoolsutils.CheckPtermDebug()

	if err := os.RemoveAll(ArtifactDir); err != nil {
		pterm.Error.Printfln("🧹 failed to delete %q: %v", ArtifactDir, err)
	} else {
		pterm.Success.Printfln("🧹 %q", ArtifactDir)
	}

	if err := os.RemoveAll(CacheDir); err != nil {
		pterm.Error.Printfln("🧹 failed to delete %q: %v", CacheDir, err)
	} else {
		pterm.Success.Printfln("🧹 %q", CacheDir)
	}

	testsOutput := filepath.Join("tests", "output")
	if err := os.RemoveAll(testsOutput); err != nil {
		pterm.Error.Printfln("🧹 failed to delete %q: %v", testsOutput, err)
	} else {
		pterm.Success.Printfln("🧹 %q", testsOutput)
	}

	pterm.Info.Println("🧹 Clean() completed")
}

// 🧪 Test runs unit and sanity tests.
func Test() { mg.SerialDeps(TestUnit, TestSanity) }

// 🧪 TestSanity runs sanity tests in containers.
func TestSanity() error {
	magetoolsutils.CheckPtermDebug()

	pterm.DefaultHeader.Println("ansible-test sanity")

	if !venvBinExists("ansible-test") {
		pterm.Error.Println("run `mage init` first")
		return nil
	}

	now := time.Now()
	if err := venvRunV(
		"ansible-test", "sanity", "--docker", "--color", "yes",
		"--exclude", "vendor/", "--exclude", ".devcontainer/",
	); err != nil {
		return err
	}
	pterm.Success.Printfln("sanity tests (took: %s)", time.Since(now))
	return nil
}

// 🧪 TestUnit runs unit tests in containers.
func TestUnit() error {
	magetoolsutils.CheckPtermDebug()

	pterm.DefaultHeader.Println("ansible-test units")

	if !venvBinExists("ansible-test") {
		pterm.Error.Println("run `mage init` first")
		return nil
	}

	testsOutput := filepath.Join("tests", "output")

	if _, err := os.Stat(testsOutput); err == nil {
		pterm.DefaultSection.Println("Cleanup old output:")
		if err := os.RemoveAll(testsOutput); err != nil {
			pterm.Error.Printfln("🧹 failed to delete %q: %v", testsOutput, err)
			return nil
		}
		pterm.Success.Printfln("🧹 %q", testsOutput)
	}

	pterm.DefaultSection.Println("Unit Tests:")

	now := time.Now()
	if err := venvRunV(
		"ansible-test", "units", "--docker", "--color", "yes", "--coverage",
	); err != nil {
		return err
	}

	pterm.Success.Printfln("unit tests (took: %s)", time.Since(now))

	pterm.DefaultSection.Println("Code Coverage Report:")

	if err := venvRun(
		"ansible-test", "coverage", "xml", "-v", "--requirements",
		"--group-by", "command", "--group-by", "version",
	); err != nil {
		return err
	}
	return venvRunV("ansible-test", "coverage", "report")
}

// work-in-progess
func Changelog() error {
	magetoolsutils.CheckPtermDebug()

	pterm.DefaultHeader.Println("antsibull-changelog")

	if !venvBinExists("pip3") {
		pterm.Error.Println("run `mage init` first")
		return nil
	}
	if !venvBinExists("antsibull-changelog") {
		if err := venvInstall("antsibull-changelog"); err != nil {
			return err
		}
	}

	return nil
}

// 🍬 Bump increments version in the galaxy file of the collection, using yq.
// Valid types are "major", "minor", "patch"
func Bump(bumpType string) error {
	pterm.DefaultHeader.Printfln("Version Bump")

	galaxyYaml := "galaxy.yml"
	current, err := sh.Output("yq", ".version", galaxyYaml)
	if err != nil {
		pterm.Error.Printfln("failed to get version from galaxy.yml:\n\t%v", err)
		return err
	}
	current = strings.TrimSpace(current)
	version, err := semver.StrictNewVersion(current)
	if err != nil {
		return err
	}

	var newVersion semver.Version
	switch bumpType {
	case "major":
		newVersion = version.IncMajor()
	case "minor":
		newVersion = version.IncMinor()
	case "patch":
		newVersion = version.IncPatch()
	default:
		return fmt.Errorf("unknown bump type: %s", bumpType)
	}
	bumped := newVersion.String()

	pterm.Info.Printfln("%q: %q -> %q", bumpType, current, bumped)

	err = sh.RunV(
		"yq", "--inplace", fmt.Sprintf(".version = \"%s\"", bumped), galaxyYaml,
	)

	if err != nil {
		pterm.Error.Printfln("failed to bump version:\n\t%v", err)
		return err
	}
	return nil
}

// 🍬 Build packages the collection into a publishable archive.
func Build() error {
	magetoolsutils.CheckPtermDebug()

	pterm.DefaultHeader.Println("ansible-galaxy collection build")

	if !venvBinExists("ansible-galaxy") {
		pterm.Error.Println("run `mage init` first")
		return nil
	}

	if err := venvRun(
		"ansible-galaxy", "collection", "build", "-v", "--force",
		"--output-path", filepath.Join(ArtifactDir, ""),
	); err != nil {
		return err
	}

	path, err := archiveFind("delinea-core*.tar.gz")
	if err != nil {
		return err
	}
	files, err := archiveContent(path)
	if err != nil {
		return err
	}

	pterm.Info.Printfln("%q:\n\t- %s", path, strings.Join(files, "\n\t- "))
	return nil
}

// 🍬 Publish sends archived collection to Ansible Galaxy.
func Publish() error {
	magetoolsutils.CheckPtermDebug()

	pterm.DefaultHeader.Println("ansible-galaxy collection publish")

	if !venvBinExists("ansible-galaxy") {
		pterm.Error.Println("run `mage init` first")
		return nil
	}

	gxServer, gxKey := os.Getenv("GALAXY_SERVER"), os.Getenv("GALAXY_KEY")
	if gxServer == "" {
		pterm.Error.Printfln("env variable `GALAXY_SERVER` is required, but not set. Skipping publish.")
		return fmt.Errorf("missing required environment variables")
	}
	if gxKey == "" {
		pterm.Error.Printfln("env variable `GALAXY_KEY` is required, but not set. Skipping publish.")
		return fmt.Errorf("missing required environment variables")
	}

	path, err := archiveFind("delinea-core*.tar.gz")
	if err != nil {
		pterm.Error.Println("run `mage build` first")
		return err
	}

	pterm.DefaultSection.Printfln("Publishing `%s` to %s", path, gxServer)

	now := time.Now()
	if err := venvRunV(
		"ansible-galaxy", "collection", "publish", "-v",
		"--server", gxServer, "--api-key", gxKey, path,
	); err != nil {
		return fmt.Errorf("running `ansible-galaxt collection publish` failed")
	}
	pterm.Success.Printfln("Published collection (took: %s)", time.Since(now))
	return nil
}

// ----------------------------------- //
//          Helper Functions           //
// ----------------------------------- //

func ansibleInit(version string) error {
	magetoolsutils.CheckPtermDebug()

	pterm.DefaultHeader.Printfln("Ansible %s Init()", AnsibleLatest)

	link := fmt.Sprintf("https://github.com/ansible/ansible/archive/%s.tar.gz", version)

	mg.SerialDeps(
		venvInit,
		func() error { return venvInstall("wheel") },
		func() error { return venvInstall(link) },
	)
	return nil
}

func venvInit() error {
	if err := mkdir(CacheDir); err != nil {
		return err
	}

	path := filepath.Join(CacheDir, "venv")
	err := sh.Run("python3", "-m", "venv", "--clear", path)
	if err != nil {
		pterm.Error.Printfln("error creating a new virtual environment: %s", err)
		return err
	}

	pterm.Success.Printfln("created a new virtual environment: %s", path)
	return nil
}

func venvBinExists(name string) bool {
	_, err := os.Stat(filepath.Join(CacheDir, "venv", "bin", name))
	return err == nil
}

func venvInstall(name string) error {
	now := time.Now()
	if err := venvRun("pip3", "install", name, "--disable-pip-version-check"); err != nil {
		pterm.Error.Printfln("error installing name %q: %s", name, err)
		return err
	}
	pterm.Success.Printfln("installed %q (took: %s)", name, time.Since(now))
	return nil
}

func venvRun(cmd string, args ...string) error  { return venvRunBinary(false, cmd, args...) }
func venvRunV(cmd string, args ...string) error { return venvRunBinary(true, cmd, args...) }

func venvRunBinary(useStdout bool, cmd string, args ...string) error {
	path := filepath.Join(CacheDir, "venv")
	venvBin := filepath.Join(path, "bin")
	runnable := filepath.Join(venvBin, cmd)

	env := map[string]string{
		"PATH":        venvBin + ":" + os.Getenv("PATH"),
		"VIRTUAL_ENV": path,
	}

	if useStdout {
		return sh.RunWithV(env, runnable, args...)
	}
	return sh.RunWith(env, runnable, args...)
}

func writeFile(path string, data string) error {
	const permBits = 0o777
	return os.WriteFile(path, []byte(data), permBits)
}

func mkdir(path string) error {
	const permBits = 0o755
	return os.MkdirAll(path, permBits)
}

func archiveFind(pattern string) (string, error) {
	archivePattern := filepath.Join(ArtifactDir, pattern)
	archives, err := filepath.Glob(archivePattern)
	if err != nil {
		return "", err
	}

	switch {
	case len(archives) == 0:
		return "", fmt.Errorf("no archive found with pattern %q", archivePattern)

	case len(archives) > 1:
		return "", fmt.Errorf("more than one archive found with pattern %q", archivePattern)

	default:
		return archives[0], nil
	}
}

func archiveContent(path string) ([]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	files := []string{}
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, header.Name)
	}
	return files, nil
}

/*
const (
	// collectionName is the name of the Ansible collection.
	collectionName = "delinea.core"

	// Namespace is the ansible collection namespace.
	Namespace = "delinea"

	// Collection is the ansible collection name.
	Collection = "core"

	// VenvToolingDirectory is the venv for tooling.
	VenvToolingDirectory = "tooling"

	// changelogFragments is the directory to store user created changelog fragments.
	changelogFragments = "changelogs/fragments"

	// GalaxyYaml is the name of the galaxy.yml file.
	GalaxyYaml = "galaxy.yml"

	// VenvDir is the default directory for virtual environment.
	VenvDir = "venv"
)

// AnsibleVersions is a list of Ansible versions to test and create virtual environments for.
var AnsibleVersions = []string{
	"stable-2.10",
	"stable-2.11",
	"stable-2.12",
	"stable-2.13",
	"devel",
}

func checklinux() {
	if runtime.GOOS == "windows" {
		_ = mg.Fatalf(1, "this command is only supported on Linux or Darwin and you are on: %s", runtime.GOOS)
	}
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
	versionSemver, _ := pterm.DefaultInteractiveTextInput.WithMultiLine(false).Show()
	pterm.Info.Printfln("You answered: %s", versionSemver)

	changelogFragmentFile := filepath.Join(changelogFragments, versionSemver+".yml")

	pterm.Info.Println("Enter release summary")
	releaseNotes, _ := pterm.DefaultInteractiveTextInput.WithMultiLine(true).Show()
	pterm.Info.Printfln("You answered: %s", releaseNotes)

	if err := writeFile(changelogFragmentFile, "---\nrelease_summary:\n    "+releaseNotes); err != nil {
		return err
	}

	if err := Venv{}.New(); err != nil {
		pterm.Error.Printfln("error installing requirements: %s", err)
		return err
	}
	pterm.Success.Println("initialized virtual environment")
	if err := venvInstall("antsibull-changelog"); err != nil {
		pterm.Error.Printfln("error installing antsibull-changelog: %v", err)
		return err
	}
	pterm.Success.Println("installed antsibull-changelog")

	if err := venvRun("antsibull-changelog", "release"); err != nil {
		return err
	}
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
func checkEnvVar(ckv *checkEnv) (string, pterm.TableData, error) {
	value, ok := os.LookupEnv(ckv.Name)
	switch {
	case ok && ckv.IsSecret:
		tbl := append(ckv.Tbl, []string{"✅", ckv.Name, "***** secret set, but not logged *****", ckv.Notes})
		return value, tbl, nil

	case ok && !ckv.IsSecret:
		tbl := append(ckv.Tbl, []string{"✅", ckv.Name, value, ckv.Notes})
		return value, tbl, nil

	case !ok && ckv.IsRequired:
		tbl := append(ckv.Tbl, []string{"❌", ckv.Name, "", ckv.Notes})
		return "", tbl, fmt.Errorf("%s is required and not set", ckv.Name)

	case !ok && !ckv.IsRequired:
		tbl := append(ckv.Tbl, []string{"👉", ckv.Name, "", ckv.Notes})
		return "", tbl, nil

	default:
		return "", nil, nil // Unreachable.
	}
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
	_, tbl, err := checkEnvVar(&checkEnv{Name: "GALAXY_SERVER", IsSecret: false, IsRequired: true, Tbl: tbl, Notes: "required for defining target publish location"})
	if err != nil {
		errorCount++
	}
	_, tbl, err = checkEnvVar(&checkEnv{Name: "GALAXY_KEY", IsSecret: true, IsRequired: true, Tbl: tbl, Notes: "required for publishing"})
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
*/
