package main

import (
	"context"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
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
	for _, f := range fileNames {
		fileDescriptor, err := fs.Add(ctx, f, mediaType, "")
		if err != nil {
			panic(err)
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)
	}

	// 2. Pack the files and tag the packed manifest
	// Note:
	// This will pack an artifact manifest by default.
	// If it does not work, try oras.PackOptions{PackImageManifest: true}.
	artifactType := "example/files"
	manifestDescriptor, err := oras.Pack(ctx, fs, artifactType, fileDescriptors, oras.PackOptions{})
	if err != nil {
		panic(err)
	}
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
	repo.Client = &auth.Client{
		Cache: auth.DefaultCache,
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
	_, err = oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		panic(err)
	}
}
