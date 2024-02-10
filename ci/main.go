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
			WithDirectory("/src", client.Host().Directory("."))
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

	new_ver := old_ver
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

		new_ver = fmt.Sprintf("%s-dev.%s",old_ver, commit_count)
	}

	out := strings.TrimSpace(new_ver)
	return (new_ver != old_ver), out, nil
}

func gen_changelog(cctx cicontext, version string) (string, error) {
	fmt.Printf("Generating Changelog for version:%s\n", version)

	convco := cctx.client.Container().From("convco/convco")
	tagged := cctx.source.
		WithExec([]string{"git", "tag", "-a", fmt.Sprintf("v%s", version), "-m", "Temp version"})

	convco = convco.
		WithDirectory("/src", tagged.Directory("/src")).
		WithWorkdir("/src")

	out, err := convco.WithExec([]string{"changelog", fmt.Sprintf("v%s", version)}).Stdout(cctx.ctx)
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
