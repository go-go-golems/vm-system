//go:build embed

package web

import (
	"embed"
	"fmt"
	"io/fs"
)

// embeddedPublic includes the generated frontend assets copied into embed/public.
//
//go:embed embed/public
var embeddedPublic embed.FS

func PublicFS() (fs.FS, error) {
	public, err := fs.Sub(embeddedPublic, "embed/public")
	if err != nil {
		return nil, fmt.Errorf("load embedded public fs: %w", err)
	}

	return public, nil
}
