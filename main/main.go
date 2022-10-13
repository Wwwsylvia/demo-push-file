package main

import (
	"context"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func main() {
	pushFiles()
	// pullFiles()
}

func pushFiles() {
	files := []string{"/tmp/myfile"}
	tag := "latest"
	fs := file.New("/tmp/")
	defer fs.Close()

	// 1. add files to a file store
	ctx := context.Background()
	desc, err := fs.PackFiles(ctx, files)
	if err != nil {
		panic(err)
	}
	err = fs.Tag(ctx, desc, tag)
	if err != nil {
		panic(err)
	}

	// 2. copy from the file store to a remote repository
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
	_, err = oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		panic(err)
	}
}

func pullFiles() {
	// 1. create a file store
	fs := file.New("")
	defer fs.Close()

	// 2. copy from a remote repository to the file store
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

	tag := "latest"
	_, err = oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		panic(err)
	}
}
