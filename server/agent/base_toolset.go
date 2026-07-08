// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"fmt"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"

	toolsartifacts "github.com/achetronic/magec/server/agent/tools/artifacts"
)

// baseToolset bundles built-in tools (like artifacts) that are injected into
// every agent by default, enabling standard capabilities across the board.
type baseToolset struct {
	tools         []tool.Tool
	artifactTools *toolsartifacts.Toolset
}

// newBaseToolset creates a new baseToolset containing the artifact tools.
//
// tempDirProvider returns the absolute path that artifact-export uses for
// transient files. The caller is responsible for resolving it (typically via
// Store.ResolveTemporaryDir, the single source of truth for that location).
//
// urlBuilder mints ephemeral signed URLs for artifacts. May be nil; when nil
// the get_artifact_url tool is not registered.
func newBaseToolset(tempDirProvider func() string, urlBuilder toolsartifacts.ArtifactURLBuilder) (*baseToolset, error) {
	artifactTs, err := toolsartifacts.NewToolset(tempDirProvider, urlBuilder)
	if err != nil {
		return nil, fmt.Errorf("failed to create artifact toolset: %w", err)
	}

	return &baseToolset{
		tools:         []tool.Tool{},
		artifactTools: artifactTs,
	}, nil
}

func (b *baseToolset) Name() string {
	return "base_toolset"
}

func (b *baseToolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	artTools, err := b.artifactTools.Tools(ctx)
	if err != nil {
		return b.tools, nil
	}
	all := make([]tool.Tool, 0, len(b.tools)+len(artTools))
	all = append(all, b.tools...)
	all = append(all, artTools...)
	return all, nil
}
