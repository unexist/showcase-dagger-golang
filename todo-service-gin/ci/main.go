//
// @package Showcase-Dagger-Golang
//
// @file Todo build
// @copyright 2023-present Christoph Kappel <christoph@unexist.dev>
// @version $Id$
//
// This program can be distributed under the terms of the Apache License v2.0.
// See the file LICENSE for details.
//

package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"dagger.io/dagger"
)

func failIfUnset(keys []string) {
	for _, key := range keys {
		if _, ok := os.LookupEnv(key); !ok {
			panic(fmt.Sprintf("$%s must be set", key))
		}
	}
}

func getEnvOrDefault(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func initClient(ctx *context.Context) (*dagger.Client, error) {
	client, err := dagger.Connect(*ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	ctx := context.Background()

	client, err := initClient(&ctx)
	if nil != err {
		panic(err)
	}

	defer client.Close()

	build(&ctx, client)

	if "1" == os.Getenv("DAGGER_PUBLISH") {
		publish(&ctx, client)
	}
}

func build(ctx *context.Context, client *dagger.Client) error {
	fmt.Println("Building with Dagger")

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci/"},
	})

	const buildPath = "build/"

	golang := client.
		Pipeline("Build application").
		Container().
		From(getEnvOrDefault("DAGGER_BUILD_IMAGE", "docker.io/golang:latest")).
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "-o",
			path.Join(buildPath, getEnvOrDefault("BINARY_NAME", "showcase"))})

	output := golang.Directory(buildPath)

	_, err := output.Export(*ctx, buildPath)
	if nil != err {
		return err
	}

	return nil
}

func WithCustomRegistryAuth(client *dagger.Client) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		_, exists := os.LookupEnv("DAGGER_REGISTRY_TOKEN")
		if exists {
			return container.WithRegistryAuth(os.Getenv("DAGGER_REGISTRY_URL"),
				os.Getenv("DAGGER_REGISTRY_USER"),
				client.SetSecret("REGISTRY_TOKEN", os.Getenv("DAGGER_REGISTRY_TOKEN")))
		}
		return container
	}
}

func WithCustomContainerByCode(client *dagger.Client) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			From(getEnvOrDefault("DAGGER_RUN_IMAGE", "docker.io/alpine:latest")).
			WithDirectory("/build", client.Host().Directory("build")).
			WithExec([]string{"mkdir", "-p", "/app"}).
			WithExec([]string{"cp", fmt.Sprintf("/build/%s",
				getEnvOrDefault("BINARY_NAME", "showcase")), "/app"}).
			WithWorkdir("/app").
			WithExposedPort(8080).
			WithDefaultTerminalCmd([]string{fmt.Sprintf("./%s",
				getEnvOrDefault("BINARY_NAME", "showcase"))})
	}
}

func WithCustomContainerByFile(client *dagger.Client) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return client.
			Host().
			Directory(".").
			DockerBuild(dagger.DirectoryDockerBuildOpts{
				Dockerfile: "./ci/Containerfile.dagger",
				BuildArgs: []dagger.BuildArg{
					{Name: "DAGGER_RUN_IMAGE", Value: getEnvOrDefault("DAGGER_RUN_IMAGE", "docker.io/alpine:latest")},
					{Name: "BINARY_NAME", Value: getEnvOrDefault("BINARY_NAME", "showcase")},
				},
			})
	}
}

func publish(ctx *context.Context, client *dagger.Client) {
	fmt.Println("Publishing with Dagger")
	failIfUnset([]string{"DAGGER_REGISTRY_URL", "DAGGER_IMAGE", "DAGGER_TAG"})

	_, err := client.
		Pipeline("Publish to Gitlab").
		Container().
		With(WithCustomRegistryAuth(client)).
		With(WithCustomContainerByCode(client)).
		// With(WithCustomContainerByFile(client)).
		Publish(*ctx,
			fmt.Sprintf("%s/root/showcase-dagger-golang/%s:%s",
				os.Getenv("DAGGER_REGISTRY_URL"),
				os.Getenv("DAGGER_IMAGE"),
				os.Getenv("DAGGER_TAG"),
			))

	if nil != err {
		fmt.Println(err)
		panic(err)
	}
}
