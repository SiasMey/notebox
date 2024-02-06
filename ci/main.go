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
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	if err := lint(context.Background(), client); err != nil {
		fmt.Println(err)
	}
	if err := test(context.Background(), client); err != nil {
		fmt.Println(err)
	}
	version, err := version(context.Background(), client)
	if err != nil {
		fmt.Println(err)
	}
	log, err := changelog(context.Background(), client, version)
	if err != nil {
		fmt.Println(err)
	}
	if err := build(context.Background(), client); err != nil {
		fmt.Println(err)
	}
	if err := publish(context.Background(), client, version, log); err != nil {
		fmt.Println(err)
	}
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

func version(ctx context.Context, client *dagger.Client) (string, error) {
	//todo(siasmey@gmail.com): This should generate and export the version
	//Version should be raised and tagged by action runner
	//Version would ideally be metadata, and not repo changes
	fmt.Println("Versioning with Dagger")

	src := client.Host().Directory(".")
	convco := client.Container().From("convco/convco")
	convco = convco.
		WithDirectory("/src", src).
		WithWorkdir("/src")

	out, err := convco.WithExec([]string{"version", "--bump"}).Stdout(ctx)
	out = strings.TrimSpace(out)

	if err != nil {
		return "", err
	}

	return out, nil
}

func changelog(ctx context.Context, client *dagger.Client, version string) (string, error) {
	// todo(siasmey@gmail.com): This should generate and export the changelog
	//Changelog should be commited by action runner
	//Changelog would ideally be metadata, and not repo commits
	fmt.Println("Changelog Generation with Dagger")

	convco := client.Container().From("convco/convco")

	//Dont know how deep this clones, might run into missing tags
	project := client.Git("https://github.com/SiasMey/notebox",
		dagger.GitOpts{KeepGitDir: true}).Branch("trunk").Tree()

	source := client.Container().From("alpine:latest").
		WithExec([]string{"apk", "add", "git"}).
		WithWorkdir("/src").
		WithDirectory("/src", project).
		WithExec([]string{"git", "tag", fmt.Sprintf("v%s", version)})

	convco = convco.
		WithDirectory("/src", source.Directory("/src")).
		WithWorkdir("/src")

	out, err := convco.WithExec([]string{"changelog"}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return out, nil
}

func publish(ctx context.Context, client *dagger.Client, version string, log string) error {
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
	return nil
}
