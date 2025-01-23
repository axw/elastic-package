// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/elastic/elastic-package/internal/cobraext"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const pullLongDescription = `Use this command to pull a package from a remote repository.

TODO more of an explanation on when/how/why to use this.`

func pullCommandAction(cmd *cobra.Command, args []string) error {
	return pullPackage(cmd.Context(), args[0], args[1])
}

func pullPackage(ctx context.Context, ref, destdir string) error {
	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return err
	}
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return err
	}
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(credStore),
	}

	fs, err := file.New(destdir)
	if err != nil {
		return err
	}
	defer fs.Close()

	if _, err := oras.Copy(ctx, repo, repo.Reference.Reference, fs, "", oras.DefaultCopyOptions); err != nil {
		return err
	}
	return nil
}

func setupPullCommand() *cobraext.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull a remote package",
		Long:  pullLongDescription,
		Args:  cobra.ExactArgs(2),
		RunE:  pullCommandAction,
	}
	return cobraext.NewCommand(cmd, cobraext.ContextGlobal)
}
