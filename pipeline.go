// Template

package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func main() {
	ctx := context.Background()

	if err := Pipeline(ctx); err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
}

func Pipeline(ctx context.Context) error {
	client, err := dagger.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Only copy what the pipeline needs into the container.
	// data/ and models/ are not on the list, so the container
	// always starts from a fresh machine - the pipeline must
	// build everything itself.
	project := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			"source_code",
			"pyproject.toml",
			"uv.lock",
			".dvc",
			".dvcignore",
			"data/images/metadata.csv.dvc",
			"data/test/metadata.csv.dvc",
		},
	})

	// Base container: uv image with your Python version,
	// with dependencies installed from the lockfile
	python := client.Container().
		From("ghcr.io/astral-sh/uv:python3.13-bookworm").
		WithDirectory("/app", project).
		WithWorkdir("/app").
		WithExec([]string{"uv", "sync"})

	// Stage 1: Data
	data := python.WithExec([]string{"uv", "run", "python", "-m", "source_code.dataset"})
	if _, err := data.Directory("data").Export(ctx, "data"); err != nil {
		return err
	}

	// Stage 2: Train
	train := data.WithExec([]string{"uv", "run", "python", "-m", "source_code.modeling.train"})
	if _, err := train.Directory("artifacts").Export(ctx, "artifacts"); err != nil {
		return err
	}

    // Stage 3: Model Selection
    selection := train.WithExec([]string{"uv", "run", "python", "-m", "source_code.modeling.model_selection"})
    if _, err := selection.Directory("models").Export(ctx, "models"); err != nil {
        return err
    }

	// Stage 4: Deployment
	deploy := train.WithExec([]string{"uv", "run", "python", "-m", "source_code.modeling.deploy"})
	// we dont save anythin so using Sync here ()
    if _, err := deploy.Sync(ctx); err!=nil {
        return err
    }

	return nil
}
