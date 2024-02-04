package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stderr))
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	if err := test(context.Background(), client); err != nil {
		fmt.Println(err)
	}
	if err := build(context.Background(), client); err != nil {
		fmt.Println(err)
	}
	if err := version(context.Background(), client); err != nil {
		fmt.Println(err)
	}
	if err := publish(context.Background(), client); err != nil {
		fmt.Println(err)
	}
}

func test(ctx context.Context, client* dagger.Client) error {
	fmt.Println("Testing with Dagger")
	return nil
}

func build(ctx context.Context, client* dagger.Client) error {
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

func version(ctx context.Context, client* dagger.Client) error {
	fmt.Println("Versioning with Dagger")

	src := client.Host().Directory(".")
	convco := client.Container().From("convco/convco")
	convco = convco.
		WithDirectory("/tmp", src).
		WithWorkdir("/tmp")

	// out, err := convco.WithExec([]string{"version", "--bump"}).Stdout(ctx)
	out, err := convco.WithExec([]string{"version", "--bump"}).Stdout(ctx)
	if err != nil {
		return err
	}
	fmt.Println(out)
	if err != nil {
		return err
	}

	out, err = convco.WithExec([]string{"changelog", "-u", out}).Stdout(ctx)
	if err != nil {
		return err
	}
	fc, err := os.Create("CHANGELOG.md")
	defer fc.Close()

	_, err = fc.WriteString(out)
	if err != nil {
		return err
	}

	return nil
}

func publish(ctx context.Context, client* dagger.Client) error {
	fmt.Println("Publishing with Dagger")
	return nil
}
