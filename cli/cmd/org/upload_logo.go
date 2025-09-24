package org

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func UploadLogoCmd(ch *cmdutil.Helper) *cobra.Command {
	var path string
	var remove bool

	cmd := &cobra.Command{
		Use:   "upload-logo [<org-name> [<path-to-image>]]",
		Args:  cobra.MaximumNArgs(2),
		Short: "Upload a custom logo",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Parse positional args into flags
			if len(args) > 0 {
				ch.Org = args[0]
				if len(args) > 1 {
					path = args[1]
				}
			}
			if ch.Org == "" {
				return fmt.Errorf("an organization name is required")
			}

			// Handle --remove
			if remove {
				if path != "" {
					return fmt.Errorf("cannot specify both --remove and a path")
				}

				// Confirmation prompt
				if ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("You are removing the custom logo for %q. Continue?", ch.Org), "", false); err != nil || !ok {
					return err
				}

				empty := ""
				_, err = client.UpdateOrganization(cmd.Context(), &adminv1.UpdateOrganizationRequest{
					Org:         ch.Org,
					LogoAssetId: &empty,
				})
				if err != nil {
					return err
				}

				ch.PrintfSuccess("Removed logo from organization %q\n", ch.Org)
				return nil
			}

			// Check the file is an image
			ext := strings.TrimPrefix(filepath.Ext(path), ".")
			switch ext {
			case "png", "jpg", "jpeg", "gif", "svg", "ico":
			default:
				return fmt.Errorf("invalid file type %q (expected PNG, JPG, GIF, SVG)", ext)
			}

			// Validate and open the path
			fi, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to read %q: %w", path, err)
			}
			if fi.IsDir() {
				return fmt.Errorf("failed to upload %q: the path is a directory", path)
			}
			if fi.Size() == 0 {
				return fmt.Errorf("failed to upload %q: the file is empty", path)
			}
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open %q: %w", path, err)
			}
			defer f.Close()

			// Confirmation prompt
			if ok, err := cmdutil.ConfirmPrompt(fmt.Sprintf("You are changing the custom logo for %q. Continue?", ch.Org), "", false); err != nil || !ok {
				return err
			}

			// Generate the asset upload URL
			asset, err := client.CreateAsset(cmd.Context(), &adminv1.CreateAssetRequest{
				Org:                ch.Org,
				Type:               "image",
				Name:               "logo",
				Extension:          ext,
				Public:             true,
				EstimatedSizeBytes: fi.Size(),
			})
			if err != nil {
				return err
			}

			// Execute the upload
			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPut, asset.SignedUrl, f)
			if err != nil {
				return fmt.Errorf("failed to upload: %w", err)
			}
			for k, v := range asset.SigningHeaders {
				req.Header.Set(k, v)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to upload: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to upload: status=%d, error=%s", resp.StatusCode, string(body))
			}

			// Update the logo
			_, err = client.UpdateOrganization(cmd.Context(), &adminv1.UpdateOrganizationRequest{
				Org:         ch.Org,
				LogoAssetId: &asset.AssetId,
			})
			if err != nil {
				return fmt.Errorf("failed to update: %w", err)
			}

			// Print confirmation message
			ch.PrintfSuccess("Updated the logo for %q\n", ch.Org)
			return nil
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization name")
	cmd.Flags().StringVar(&path, "path", "", "Path to image file (PNG or JPEG)")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove the current logo")

	return cmd
}
