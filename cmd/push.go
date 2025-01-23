// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cmd

import (
	"fmt"

	"github.com/elastic/elastic-package/internal/cobraext"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const pushLongDescription = `Use this command to push a package to a remote repository.

TODO more of an explanation on when/how/why to use this.`

func pushCommandAction(cmd *cobra.Command, args []string) error {
	fileStore, err := file.New(args[0])
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fileStore.Close()

	// 1. Add the package directory as a layer to the file store.
	fileDescriptor, err := fileStore.Add(cmd.Context(), ".", "", "")
	if err != nil {
		return err
	}

	// 2. Pack the files and tag the packed manifest
	artifactType := "application/vnd.elastic.package"
	opts := oras.PackManifestOptions{
		Layers: []v1.Descriptor{fileDescriptor},
	}
	sourceManifestDescriptor, err := oras.PackManifest(
		cmd.Context(), fileStore,
		oras.PackManifestVersion1_1,
		artifactType, opts,
	)
	if err != nil {
		return err
	}

	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return err
	}
	repo, err := remote.NewRepository(args[1])
	if err != nil {
		return err
	}
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(credStore),
	}

	tag := repo.Reference.Reference
	if err := fileStore.Tag(cmd.Context(), sourceManifestDescriptor, tag); err != nil {
		return err
	}
	if _, err := oras.Copy(
		cmd.Context(),
		fileStore,
		repo.Reference.Reference,
		repo,
		"",
		oras.DefaultCopyOptions,
	); err != nil {
		return err
	}
	return nil
}

func setupPushCommand() *cobraext.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push a remote package",
		Long:  pushLongDescription,
		Args:  cobra.ExactArgs(2),
		RunE:  pushCommandAction,
	}
	return cobraext.NewCommand(cmd, cobraext.ContextGlobal)
}
