import {
  createRuntimeServiceListResources,
  type V1Message,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import {
  createFileDiffBlock,
  type FileDiffBlock,
} from "@rilldata/web-common/features/chat/core/messages/file-diff/file-diff-block.ts";
import { derived } from "svelte/store";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

// =============================================================================
// BLOCK TYPE
// =============================================================================

export type DevelopBlock = {
  type: "develop";
  id: string;
  diffs: FileDiffBlock[];
  generatedResources: V1ResourceName[];
  checkpointCommitHash: string;
};

/**
 * Creates a file diff block from a write_file tool call message.
 * Returns null if the data is invalid or the result indicates an error.
 */
export function createDevelopBlock(
  writeMessages: V1Message[],
  id: string,
  resultMessagesByParentId: Map<string | undefined, V1Message>,
): DevelopBlock | null {
  try {
    const diffs = writeMessages.map((message) =>
      createFileDiffBlock(message, resultMessagesByParentId.get(message.id)),
    );

    const generatedResources: V1ResourceName[] = [];
    const resourceNamesSeen = new Set<string>();
    diffs.forEach((diff) => {
      if (!diff) return;
      diff.generatedResources.forEach(({ kind, name }) => {
        const key = kind + "__" + name;
        if (resourceNamesSeen.has(key)) return;
        resourceNamesSeen.add(key);
        generatedResources.push({ kind, name });
      });
    });

    return {
      type: "develop",
      id,
      diffs: diffs.filter(Boolean) as FileDiffBlock[],
      generatedResources,
      checkpointCommitHash: diffs[0]?.checkpointCommitHash || "",
    };
  } catch {
    return null;
  }
}

export function getGenerateCTAs(instanceId: string, block: DevelopBlock) {
  const resourcesQuery = createRuntimeServiceListResources(
    instanceId,
    undefined,
  );
  return derived(resourcesQuery, (resourcesResp) => {
    const resources = resourcesResp.data?.resources ?? [];
    const models: string[] = [];
    const metricsViews: string[] = [];

    block.generatedResources.forEach(({ kind, name }) => {
      const hasResource = resources.find(
        (r) => r.meta?.name?.kind === kind && r.meta?.name?.name === name,
      );
      if (!hasResource) return; // resource was deleted

      switch (kind) {
        case ResourceKind.Model: {
          const hasSomeMetricsView = resources.find(
            (r) =>
              r.metricsView?.spec?.model === name ||
              r.metricsView?.spec?.table === name,
          );
          if (!hasSomeMetricsView) models.push(name!);
          break;
        }

        case ResourceKind.MetricsView: {
          const hasSomeExplore = resources.find(
            (r) => r.explore?.spec?.metricsView === name,
          );
          if (!hasSomeExplore) metricsViews.push(name!);
          break;
        }
      }
    });

    return {
      models,
      metricsViews,
    };
  });
}
