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

	"dagger.io/dagger"
)

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
	publish(&ctx, client)
}

func build(ctx *context.Context, client *dagger.Client) error {
	fmt.Println("Building with Dagger")

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci/"},
	})

	const path = "build/"

	golang := client.
		Pipeline("Build application").
		Container().
		From("golang:latest").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "-o", path})

	output := golang.Directory(path)

	_, err := output.Export(*ctx, path)
	if nil != err {
		return err
	}

	return nil
}

func publish(ctx *context.Context, client *dagger.Client) {
	imgPath := fmt.Sprintf("%s/%s:%s",
		os.Getenv("USERNAME"),
		os.Getenv("IMAGE_NAME"),
		os.Getenv("TAG"),
	)

	_, err := client.
		Pipeline("Publish to Docker Hub").
		Host().
		Directory(".").
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: "./Dockerfile.prod",
		}).
		Publish(*ctx, imgPath)

	if nil != err {
		fmt.Println(err)
		panic(err)
	}
}
