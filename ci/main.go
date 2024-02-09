package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
)

func main() {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stderr))
	if os.Getenv("GH_SECRET") == "" {
		panic("Environment variable GH_SECRET is not set")
	}
	gh_pat := client.SetSecret("gh-pat-secret", os.Getenv("GH_SECRET"))

	git_src := client.Container().From("alpine:latest").
		WithExec([]string{"apk", "add", "git"}).
		WithWorkdir("/src").
		WithSecretVariable("GH_SECRET", gh_pat).
		WithFile("/root/.gitconfig", client.Host().File("./ci/.gitconfig")).
		WithExec([]string{"git", "clone", "https://github.com/SiasMey/notebox.git", "."})

	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	version, err := version(context.Background(), client, git_src)
	if err != nil {
		fmt.Println(err)
	}
	log, err := changelog(context.Background(), client, git_src, version)
	if err != nil {
		fmt.Println(err)
	}

	if err := publish(context.Background(), client, git_src, version, log); err != nil {
		fmt.Println(err)
	}
}

func version(ctx context.Context, client *dagger.Client, git_src *dagger.Container) (string, error) {
	//todo(siasmey@gmail.com): This should generate and export the version
	//Version should be raised and tagged by action runner
	//Version would ideally be metadata, and not repo changes
	fmt.Println("Versioning with Dagger")

	convco := client.Container().From("convco/convco")
	convco = convco.
		WithDirectory("/src", git_src.Directory("/src")).
		WithWorkdir("/src")

	out, err := convco.WithExec([]string{"version", "--bump"}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	out = strings.TrimSpace(out)
	return out, nil
}

func changelog(ctx context.Context, client *dagger.Client, git_src *dagger.Container, version string) (string, error) {
	convco := client.Container().From("convco/convco")
	check, err := git_src.
		WithExec([]string{"git", "tag", "-a", fmt.Sprintf("v%s", version), "-m", "Temp version"}).
		WithExec([]string{"git", "tag"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}
	fmt.Println(check)

	convco = convco.
		WithDirectory("/src", git_src.Directory("/src")).
		WithWorkdir("/src")

	out, err := convco.WithExec([]string{"changelog"}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return out, nil
}

func publish(ctx context.Context, client *dagger.Client, git_src *dagger.Container, version string, log string) error {
	//todo(siasmey@gmail.com): publish all artifacts to platform
	//changelog/version and build artifacts need to go in here
	//should this clone, tag and commit before doing the publish?
	fmt.Println("Publishing with Dagger")

	fv, err := os.Create("version.txt")
	defer fv.Close()
	if err != nil {
		return err
	}
	_, err = fv.WriteString(version)
	if err != nil {
		return err
	}

	fc, err := os.Create("CHANGELOG.md")
	defer fc.Close()
	if err != nil {
		return err
	}
	_, err = fc.WriteString(log)
	if err != nil {
		return err
	}

	check, err := git_src.
		WithFile("version.txt", client.Host().File("version.txt")).
		WithFile("CHANGELOG.md", client.Host().File("CHANGELOG.md")).
		WithExec([]string{"git", "add", "version.txt", "CHANGELOG.md"}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("chore: release %s [skip ci]", version)}).
		WithExec([]string{"git", "tag", "-a", fmt.Sprintf("v%s", version), "-m", "Release Version"}).
		WithExec([]string{"git", "push", "--follow-tags"}).
		Stdout(ctx)
	if err != nil {
		return err
	}
	fmt.Println(check)
	return nil
}

func lint(ctx context.Context, client *dagger.Client) error {
	//todo(siasmey@gmail.com): All the linters, export feedback as files
	//Feedback should be raised by action runner
	fmt.Println("Linting with Dagger")
	return nil
}

func test(ctx context.Context, client *dagger.Client) error {
	//todo(siasmey@gmail.com): All the Tests, export feedback as files
	//Feedback should be raised by action runner
	fmt.Println("Testing with Dagger")
	return nil
}

func build(ctx context.Context, client *dagger.Client) error {
	fmt.Println("Building with Dagger")

	// define build matrix
	oses := []string{"linux", "darwin"}
	arches := []string{"amd64", "arm64"}

	// get reference to the local project
	cmd := client.Host().Directory("./cmd")
	pkg := client.Host().Directory("./pkg")
	gomod := client.Host().File("./go.mod")
	gosum := client.Host().File("./go.sum")

	// create empty directory to put build outputs
	outputs := client.Directory()

	// get `golang` image
	golang := client.Container().From("golang:1.21")

	// mount cloned repository into `golang` image
	golang = golang.
		WithDirectory("/src/cmd", cmd).
		WithDirectory("/src/pkg", pkg).
		WithFile("/src/go.mod", gomod).
		WithFile("/src/go.sum", gosum).
		WithWorkdir("/src").
		WithMountedCache("/go/pkg/mod", client.CacheVolume("go-mod-121")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithMountedCache("/go/build-cache", client.CacheVolume("go-build-121")).
		WithEnvVariable("GOCACHE", "/go/build-cache")

	for _, goos := range oses {
		for _, goarch := range arches {
			// create a directory for each os and arch
			path := fmt.Sprintf("build/%s/%s/", goos, goarch)

			// set GOARCH and GOOS in the build environment
			build := golang.WithEnvVariable("GOOS", goos)
			build = build.WithEnvVariable("GOARCH", goarch)

			binary := fmt.Sprintf("%s/%s", path, "nbx")

			// build application
			build = build.WithExec([]string{"go", "build", "-o", binary, "cmd/nbx/main.go"})

			// get reference to build output directory in container
			outputs = outputs.WithDirectory(path, build.Directory(path))
		}
	}
	// write build artifacts to host
	_, err := outputs.Export(ctx, ".")
	if err != nil {
		return err
	}

	return nil
}
