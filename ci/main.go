package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"dagger.io/dagger"
)

func main() {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if os.Getenv("GH_SECRET") == "" {
		panic("Environment variable GH_SECRET is not set")
	}
	gh_pat := client.SetSecret("gh-pat-secret", os.Getenv("GH_SECRET"))
	is_remote := os.Getenv("GH_ACTION") != ""
	git_src, err := get_source(context.Background(), client, gh_pat)
	if err != nil {
		panic(err)
	}

	bump, version, err := version(context.Background(), client, git_src)
	if err != nil {
		panic(err)
	}
	if !bump {
		fmt.Println("No version bump, exiting pipeline")
		os.Exit(0)
	}
	log, err := changelog(context.Background(), client, git_src, version)
	if err != nil {
		panic(err)
	}

	if err := publish(context.Background(), client, git_src, version, log, is_remote); err != nil {
		panic(err)
	}
}

func get_source(ctx context.Context, client *dagger.Client, secret *dagger.Secret) (*dagger.Container, error) {
	git_src := client.Container().From("alpine:latest").
		WithExec([]string{"apk", "add", "git"}).
		WithWorkdir("/src").
		WithSecretVariable("GH_SECRET", secret).
		WithFile("/root/.gitconfig", client.Host().File("./ci/.gitconfig")).
		WithEnvVariable("CACHEBUSTER", time.Now().String()).
		WithExec([]string{"git", "clone", "https://github.com/SiasMey/notebox.git", "."})

	return git_src, nil
}

func version(ctx context.Context, client *dagger.Client, git_src *dagger.Container) (bool, string, error) {
	fmt.Println("Versioning with Dagger")

	convco := client.Container().From("convco/convco")
	convco = convco.
		WithDirectory("/src", git_src.Directory("/src")).
		WithWorkdir("/src")

	old_ver, err := convco.WithExec([]string{"version"}).Stdout(ctx)
	if err != nil {
		return false, "", err
	}
	new_ver, err := convco.WithExec([]string{"version", "--bump"}).Stdout(ctx)
	if err != nil {
		return false, "", err
	}

	out := strings.TrimSpace(new_ver)
	return (new_ver != old_ver), out, nil
}

func changelog(ctx context.Context, client *dagger.Client, git_src *dagger.Container, version string) (string, error) {
	convco := client.Container().From("convco/convco")
	tagged := git_src.
		WithExec([]string{"git", "tag", "-a", fmt.Sprintf("v%s", version), "-m", "Temp version"})

	convco = convco.
		WithDirectory("/src", tagged.Directory("/src")).
		WithWorkdir("/src")

	out, err := convco.WithExec([]string{"changelog", fmt.Sprintf("v%s", version)}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return out, nil
}

func publish(ctx context.Context, client *dagger.Client, git_src *dagger.Container, version string, log string, is_remote bool) error {
	//todo(siasmey@gmail.com): publish all artifacts to platform
	//changelog/version and build artifacts need to go in here
	//should this clone, tag and commit before doing the publish?
	fmt.Println("Publishing with Dagger")
	fmt.Println(version)
	fmt.Println(log)

	fc, err := os.CreateTemp("", "changelog")
	if err != nil {
		return err
	}
	defer os.Remove(fc.Name())
	_, err = fc.WriteString(log)
	if err != nil {
		return err
	}
	git_src = git_src.WithFile("CHANGELOG.md", client.Host().File(fc.Name()))

	check, err := git_src.
		WithExec([]string{"git", "add", "CHANGELOG.md"}).
		WithExec([]string{"git", "commit", "-m", fmt.Sprintf("chore: release %s [skip ci]", version)}).
		WithExec([]string{"git", "tag", "-a", fmt.Sprintf("v%s", version), "-m", "Release Version"}).
		Stdout(ctx)
	if err != nil {
		return err
	}
	fmt.Println(check)

	if !is_remote {
		check, err := git_src.
			WithExec([]string{"git", "push", "--follow-tags"}).
			Stdout(ctx)
		if err != nil {
			return err
		}
		fmt.Println(check)
	}
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
