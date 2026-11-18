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

    // ...
}