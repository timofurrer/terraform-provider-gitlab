package provider

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getAvatarSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"avatar": {
			Description: "A local path to the avatar image to upload. **Note**: not available for imported resources.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"avatar_hash": {
			Description:  "The hash of the avatar image. Use `filesha256(\"path/to/avatar.png\")` whenever possible. **Note**: this is used to trigger an update of the avatar. If it's not given, but an avatar is given, the avatar will be updated each time. **Note**: not available for imported resources.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			RequiredWith: []string{"avatar"},
		},
		"avatar_url": {
			Description: "The URL of the avatar image.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func avatarDiff(ctx context.Context, rd *schema.ResourceDiff, i interface{}) error {
	if _, ok := rd.GetOk("avatar"); ok {
		if v, ok := rd.GetOk("avatar_hash"); !ok || v.(string) == "" {
			if err := rd.SetNewComputed("avatar_hash"); err != nil {
				return err
			}
		}
	}
	return nil
}

type avatar struct {
	Filename string
	Image    io.Reader
}

func createAvatar(d *schema.ResourceData) (*avatar, error) {
	if v, ok := d.GetOk("avatar"); ok {
		avatarPath := v.(string)
		return readLocalAvatar(avatarPath)
	}
	return nil, nil
}

func updateAvatar(d *schema.ResourceData) (*avatar, error) {
	if d.HasChanges("avatar", "avatar_hash") || d.Get("avatar_hash").(string) == "" {
		avatarPath := d.Get("avatar").(string)
		// NOTE: the avatar should be removed
		if avatarPath == "" {
			// terraform doesn't care to remove this from state, thus, we do.
			d.Set("avatar_hash", "")
			return &avatar{}, nil
		} else {
			changedAvatar, err := readLocalAvatar(avatarPath)
			if err != nil {
				return nil, err
			}
			return changedAvatar, nil
		}
	}

	return nil, nil
}

func readLocalAvatar(avatarPath string) (*avatar, error) {
	avatarFile, err := os.Open(avatarPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open avatar file %s: %s", avatarPath, err)
	}

	return &avatar{
		Filename: avatarPath,
		Image:    avatarFile,
	}, nil
}
