package main

import (
	"context"
	"fmt"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func main() {
	pushFiles()
	pullFiles()
}

func pushFiles() {
	// 0. Create a file store
	fs, err := file.New("/tmp/")
	if err != nil {
		panic(err)
	}
	defer fs.Close()
	ctx := context.Background()

	// 1. Add files to a file store
	mediaType := "example/file"
	fileNames := []string{"/tmp/myfile"}
	fileDescriptors := make([]v1.Descriptor, 0, len(fileNames))
	for _, name := range fileNames {
		fileDescriptor, err := fs.Add(ctx, name, mediaType, "")
		if err != nil {
			panic(err)
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)
		fmt.Printf("file descriptor for %s: %v\n", name, fileDescriptor)
	}

	// 2. Pack the files and tag the packed manifest
	// Note:
	// oras.Pack() packs an artifact manifest by default.
	// If the remote repository does not support the artifact manifest media type,
	// try Image manifest instead by specifying oras.PackOptions{PackImageManifest: true}.
	artifactType := "example/files"
	manifestDescriptor, err := oras.Pack(ctx, fs, artifactType, fileDescriptors, oras.PackOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("manifest descriptor:", manifestDescriptor)

	tag := "latest"
	if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		panic(err)
	}

	// 3. Connect to a remote repository
	reg := "myregistry.example.com"
	repo, err := remote.NewRepository(reg + "/myrepo")
	if err != nil {
		panic(err)
	}
	// Note: The below code can be omitted if authentication is not required
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(reg, auth.Credential{
			Username: "username",
			Password: "password",
		}),
	}

	// 3. Copy from the file store to the remote repository
	_, err = oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		panic(err)
	}
}

func pullFiles() {
	// 0. Create a file store
	fs, err := file.New("/tmp/")
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	// 1. Connect to a remote repository
	ctx := context.Background()
	reg := "myregistry.example.com"
	repo, err := remote.NewRepository(reg + "/myrepo")
	if err != nil {
		panic(err)
	}
	repo.Client = &auth.Client{
		Cache: auth.DefaultCache,
		Credential: auth.StaticCredential(reg, auth.Credential{
			Username: "username",
			Password: "password",
		}),
	}

	// 2. Copy from the remote repository to the file store
	tag := "latest"
	manifestDescriptor, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		panic(err)
	}
	fmt.Println("manifest descriptor:", manifestDescriptor)
}
