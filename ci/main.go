package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

func main() {
	if err := test(context.Background()); err != nil {
		fmt.Println(err)
	}
	if err := build(context.Background()); err != nil {
		fmt.Println(err)
	}
	if err := version(context.Background()); err != nil {
		fmt.Println(err)
	}
	if err := publish(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func test(ctx context.Context) error {
	fmt.Println("Testing with Dagger")
	return nil
}

func build(ctx context.Context) error {
	fmt.Println("Building with Dagger")

	// define build matrix
	oses := []string{"linux", "darwin"}
	arches := []string{"amd64", "arm64"}

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

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
	_, err = outputs.Export(ctx, ".")
	if err != nil {
		return err
	}

	return nil
}

func version(ctx context.Context) error {
	fmt.Println("Versioning with Dagger")
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))

	src := client.Host().Directory(".")

	convco := client.Container().From("convco/convco")
	convco = convco.
		WithDirectory("/src", src).
		WithWorkdir("/src")

	// out, err := convco.WithExec([]string{"version", "--bump"}).Stdout(ctx)
	out, err := convco.WithExec([]string{"version", "--bump"}).Stdout(ctx)
	fmt.Println(out)

	if err != nil {
		return err
	}

	return nil
}

func publish(ctx context.Context) error {
	fmt.Println("Publishing with Dagger")
	return nil
}
