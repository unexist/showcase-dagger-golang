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

func publish(ctx *context.Context, client *dagger.Client) {
	fmt.Println("Publishing with Dagger")
	failIfUnset([]string{"DAGGER_REGISTRY_TOKEN", "DAGGER_REGISTRY_URL", "DAGGER_REGISTRY_USER", "DAGGER_IMAGE", "DAGGER_TAG"})

	_, err := client.
		Pipeline("Publish to Gitlab").
		Host().
		Directory(".").
		DockerBuild().

		// Create container by code
		From(getEnvOrDefault("DAGGER_RUN_IMAGE", "docker.io/alpine:latest")).
		WithExec([]string{"mkdir -p /app",
			fmt.Sprintf("cp build/%s /app", getEnvOrDefault("BINARY_NAME", "showcase"))},
			dagger.ContainerWithExecOpts{}).
		WithWorkdir("/app").
		WithExposedPort(8080, dagger.ContainerWithExposedPortOpts{}).
		WithDefaultTerminalCmd([]string{fmt.Sprintf("./%s", getEnvOrDefault("BINARY_NAME", "showcase"))},
			dagger.ContainerWithDefaultTerminalCmdOpts{}).

		// Create container by containerfile
		//	DockerBuild(dagger.DirectoryDockerBuildOpts{
		//		Dockerfile: "./ci/Containerfile.dagger",
		//		BuildArgs: []dagger.BuildArg{
		//			{Name: "DAGGER_RUN_IMAGE", Value: getEnvOrDefault("DAGGER_RUN_IMAGE", "docker.io/alpine:latest")},
		//			{Name: "BINARY_NAME", Value: getEnvOrDefault("BINARY_NAME", "showcase")},
		//		},
		//	}).

		WithRegistryAuth(os.Getenv("DAGGER_REGISTRY_URL"),
			os.Getenv("DAGGER_REGISTRY_USER"),
			client.SetSecret("REGISTRY_TOKEN", os.Getenv("DAGGER_REGISTRY_TOKEN"))).
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
