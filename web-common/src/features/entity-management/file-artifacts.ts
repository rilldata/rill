import {
  ResourceKind,
  fetchResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1Resource,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, get, writable } from "svelte/store";
import { FileArtifact } from "./file-artifact";

export class FileArtifacts {
  private readonly artifacts: Map<string, FileArtifact> = new Map();
  readonly unsavedFiles = writable(new Set<string>());

  async init(queryClient: QueryClient, instanceId: string) {
    const resources = await fetchResources(queryClient, instanceId);
    for (const resource of resources) {
      switch (resource.meta?.name?.kind) {
        case ResourceKind.Connector:
        case ResourceKind.Source:
        case ResourceKind.Model:
        case ResourceKind.MetricsView:
        case ResourceKind.Explore:
        case ResourceKind.Component:
        case ResourceKind.Canvas:
        case ResourceKind.Theme:
        case ResourceKind.API:
          // set query data for GetResource to avoid refetching data we already have
          queryClient.setQueryData(
            getRuntimeServiceGetResourceQueryKey(instanceId, {
              "name.name": resource.meta?.name?.name,
              "name.kind": resource.meta?.name?.kind,
            }),
            {
              resource,
            },
          );
          this.updateArtifacts(resource);
          break;
      }
    }
  }

  removeFile(filePath: string) {
    this.artifacts.delete(filePath);
  }

  deleteResource(name: V1ResourceName) {
    const artifact = this.findFileArtifact(
      (name.kind ?? "") as ResourceKind,
      name.name ?? "",
    );
    if (!artifact) return;

    this.getFileArtifact(artifact.path)?.hardDeleteResource();
  }

  updateArtifacts(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.getFileArtifact(filePath)?.updateResource(resource);
    });
  }

  getFileArtifact = (filePath: string) => {
    let artifact = this.artifacts.get(filePath);

    if (!artifact) {
      artifact = new FileArtifact(filePath);
      this.artifacts.set(filePath, artifact);
    }

    return artifact;
  };

  findFileArtifact(resKind: ResourceKind, resName: string) {
    for (const [, artifact] of this.artifacts.entries()) {
      if (!artifact) continue;
      const name = get(artifact.resourceName);
      if (name?.kind === resKind && name?.name === resName) {
        return artifact;
      }
    }
    return undefined;
  }

  /**
   * Best effort list of all reconciling resources.
   */
  getReconcilingResourceNames() {
    const artifacts = Array.from(this.artifacts.values());
    return derived(
      artifacts.map((a) => a.reconciling),
      (reconcilingArtifacts) => {
        const currentlyReconciling = new Array<V1ResourceName>();
        reconcilingArtifacts.forEach((reconcilingArtifact, i) => {
          if (reconcilingArtifact) {
            currentlyReconciling.push(
              get(artifacts[i].resourceName) as V1ResourceName,
            );
          }
        });
        return currentlyReconciling;
      },
    );
  }

  /**
   * Filters all fileArtifacts based on kind param and returns the file paths.
   * This can be expensive if the project gets large.
   * If we ever need this reactively then we should look into caching this list.
   */
  getNamesForKind(kind: ResourceKind): string[] {
    return Array.from(this.artifacts.values())
      .filter((artifact) => get(artifact.resourceName)?.kind === kind)
      .map((artifact) => get(artifact.resourceName)?.name ?? "");
  }

  async saveAll() {
    await Promise.all(
      Array.from(this.artifacts.values()).map((artifact) =>
        artifact.saveLocalContent(),
      ),
    );
  }
}

export const fileArtifacts = new FileArtifacts();
