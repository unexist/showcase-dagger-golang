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

func getEnv(key, fallback string) string {
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
		From("golang:latest").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "-o",
			path.Join(buildPath, getEnv("BINARY_NAME", "showcase"))})

	output := golang.Directory(buildPath)

	_, err := output.Export(*ctx, buildPath)
	if nil != err {
		return err
	}

	return nil
}

func publish(ctx *context.Context, client *dagger.Client) {
	fmt.Println("Publishing with Dagger")
	failIfUnset([]string{"DAGGER_REGISTRY", "DAGGER_IMAGE", "DAGGER_TAG"})

	_, err := client.
		Pipeline("Publish to Gitlab").
		Host().
		Directory(".").
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: "./ci/Containerfile.dagger",
			BuildArgs: []dagger.BuildArg{
				{Name: "BINARY_NAME", Value: getEnv("BINARY_NAME", "showcase")},
			},
		}).
		Publish(*ctx,
			fmt.Sprintf("%s/root/showcase-dagger-golang/%s:%s",
				os.Getenv("DAGGER_REGISTRY"),
				os.Getenv("DAGGER_IMAGE"),
				os.Getenv("DAGGER_TAG"),
			))

	if nil != err {
		fmt.Println(err)
		panic(err)
	}
}
