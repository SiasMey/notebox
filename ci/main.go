package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"dagger.io/dagger"
)

type cicontext struct {
	ctx       context.Context
	client    *dagger.Client
	source    *dagger.Container
	is_remote bool
}

func main() {
	cctx, err := setup()
	if err != nil {
		panic(err)
	}

	err = check_format(cctx)
	if err != nil {
		panic(err)
	}
	err = lint(cctx)
	if err != nil {
		panic(err)
	}
	bump, version, err := version(cctx)
	if err != nil {
		panic(err)
	}
	if !bump {
		fmt.Println("No version bump, exiting pipeline")
		os.Exit(0)
	}
	changelog, err := gen_changelog(cctx, version)
	if err != nil {
		panic(err)
	}
	err = build(cctx)
	if err != nil {
		panic(err)
	}

	if err := publish(cctx, version, changelog); err != nil {
		panic(err)
	}
}

func get_source(ctx context.Context, client *dagger.Client, secret *dagger.Secret, is_remote bool) (*dagger.Container, error) {
	fmt.Println("Cloning Source")
	git_src := client.Container().From("alpine:latest").
		WithExec([]string{"apk", "add", "git"}).
		WithWorkdir("/src").
		WithSecretVariable("GH_SECRET", secret).
		WithFile("/root/.gitconfig", client.Host().File("./ci/.gitconfig"))

	if is_remote {
		git_src = git_src.
			WithEnvVariable("CACHEBUSTER", time.Now().String()).
			WithExec([]string{"git", "clone", "https://github.com/SiasMey/notebox.git", "."})
	} else {
		git_src = git_src.
			WithDirectory("/src/.git", client.Host().Directory("./.git")).
			WithDirectory("/src/pkg", client.Host().Directory("./pkg")).
			WithDirectory("/src/cmd", client.Host().Directory("./cmd")).
			WithFile("/src/go.mod", client.Host().File("./go.mod")).
			WithFile("/src/go.sum", client.Host().File("./go.sum"))
	}

	return git_src, nil
}

func version(cctx cicontext) (bool, string, error) {
	fmt.Println("Versioning source")

	convco := cctx.client.Container().From("convco/convco")
	convco = convco.
		WithDirectory("/src", cctx.source.Directory("/src")).
		WithWorkdir("/src")

	old_ver, err := convco.WithExec([]string{"version"}).Stdout(cctx.ctx)
	if err != nil {
		return false, "", err
	}
	old_ver = strings.TrimSpace(old_ver)

	new_ver := ""
	if cctx.is_remote {
		new_ver, err = convco.WithExec([]string{"version", "--bump"}).Stdout(cctx.ctx)
		if err != nil {
			return false, "", err
		}
	} else {
		commit_count, err := cctx.source.WithExec([]string{"git", "rev-list", fmt.Sprintf("v%s..HEAD", old_ver), "--count"}).Stdout(cctx.ctx)
		if err != nil {
			return false, "", err
		}

		new_ver = fmt.Sprintf("%s-dev.%s", old_ver, commit_count)
	}

	out := strings.TrimSpace(new_ver)
	return (new_ver != old_ver), out, nil
}

func gen_changelog(cctx cicontext, version string) (string, error) {
	fmt.Printf("Generating Changelog for version:%s\n", version)

	var out string
	var err error

	convco := cctx.client.Container().From("convco/convco")
	if cctx.is_remote {
		tagged := cctx.source.
			WithExec([]string{"git", "tag", "-a", fmt.Sprintf("v%s", version), "-m", "Temp version"})

		convco = convco.
			WithDirectory("/src", tagged.Directory("/src")).
			WithWorkdir("/src")
		out, err = convco.WithExec([]string{"changelog", "-m", "20", fmt.Sprintf("v%s", version)}).Stdout(cctx.ctx)
	} else {
		convco = convco.
			WithDirectory("/src", cctx.source.Directory("/src")).
			WithWorkdir("/src")
		out, err = convco.WithExec([]string{"changelog", "-m", "2"}).Stdout(cctx.ctx)
	}
	if err != nil {
		return "", err
	}

	return out, nil
}

func publish(cctx cicontext, version string, changelog string) error {
	fmt.Printf("Publishing version:%s \n", version)

	fc, err := os.CreateTemp("", "changelog")
	if err != nil {
		return err
	}
	defer os.Remove(fc.Name())

	_, err = fc.WriteString(changelog)
	if err != nil {
		return err
	}
	src := cctx.source.WithFile("CHANGELOG.md", cctx.client.Host().File(fc.Name()))

	if cctx.is_remote {
		fmt.Println("Remote execution: Publishing enabled")
		_, err := src.
			WithExec([]string{"git", "add", "CHANGELOG.md"}).
			WithExec([]string{"git", "commit", "-m", fmt.Sprintf("chore: release %s [skip ci]", version)}).
			WithExec([]string{"git", "tag", "-a", fmt.Sprintf("v%s", version), "-m", "Release Version"}).
			WithExec([]string{"git", "push", "--follow-tags"}).
			Stdout(cctx.ctx)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(changelog)
	}
	return nil
}

func check_format(cctx cicontext) error {
	fmt.Println("Checking file format")
	golang := cctx.client.Container().From("golang:1.21")
	format, err := golang.
		WithWorkdir("/src").
		WithDirectory("/src", cctx.source.Directory("/src")).
		WithMountedCache("/go/pkg/mod", cctx.client.CacheVolume("go-mod-121")).
		WithMountedCache("/go/build-cache", cctx.client.CacheVolume("go-build-121")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
		WithEnvVariable("GOCACHE", "/go/build-cache").
		WithExec([]string{"gofmt", "-s", "-d", "."}).
		Stdout(cctx.ctx)
	if err != nil {
		return err
	}
	if format != "" {
		return errors.New(format)
	}
	return nil
}

func lint(cctx cicontext) error {
	fmt.Println("Linting")

	golang := cctx.client.Container().From("golangci/golangci-lint:v1.56.1")
	lint, err := golang.
		WithWorkdir("/src").
		WithDirectory("/src", cctx.source.Directory("/src")).
		WithMountedCache("/root/.cache", cctx.client.CacheVolume("golangci-lint-cache")).
		WithExec([]string{"golangci-lint", "run", "-v"}).
		Stdout(cctx.ctx)
	if err != nil {
		return err
	}
	if lint != "" {
		return errors.New(lint)
	}
	return nil
}

func build(cctx cicontext) error {
	fmt.Println("Building with Dagger")

	// define build matrix
	oses := []string{"linux", "darwin"}
	arches := []string{"amd64", "arm64"}

	// get reference to the local project
	cmd := cctx.client.Host().Directory("./cmd")
	pkg := cctx.client.Host().Directory("./pkg")
	gomod := cctx.client.Host().File("./go.mod")
	gosum := cctx.client.Host().File("./go.sum")

	// create empty directory to put build outputs
	outputs := cctx.client.Directory()

	// get `golang` image
	golang := cctx.client.Container().From("golang:1.21")

	// mount cloned repository into `golang` image
	golang = golang.
		WithDirectory("/src/cmd", cmd).
		WithDirectory("/src/pkg", pkg).
		WithFile("/src/go.mod", gomod).
		WithFile("/src/go.sum", gosum).
		WithWorkdir("/src").
		WithMountedCache("/go/pkg/mod", cctx.client.CacheVolume("go-mod-121")).
		WithMountedCache("/go/build-cache", cctx.client.CacheVolume("go-build-121")).
		WithEnvVariable("GOMODCACHE", "/go/pkg/mod").
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
	_, err := outputs.Export(cctx.ctx, ".")
	if err != nil {
		return err
	}

	return nil
}

func setup() (cicontext, error) {
	ctx := context.Background()
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return cicontext{}, err
	}
	defer client.Close()

	if os.Getenv("GH_SECRET") == "" {
		return cicontext{}, errors.New("No GH_SECRET env var set")
	}
	gh_pat := client.SetSecret("gh-pat-secret", os.Getenv("GH_SECRET"))

	is_remote := os.Getenv("GH_ACTION") != ""

	git_src, err := get_source(context.Background(), client, gh_pat, is_remote)
	if err != nil {
		return cicontext{}, err
	}

	return cicontext{ctx: ctx, client: client, source: git_src, is_remote: is_remote}, nil
}
