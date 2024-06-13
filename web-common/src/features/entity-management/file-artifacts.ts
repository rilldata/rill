import { parseKindAndNameFromFile } from "@rilldata/web-common/features/entity-management/file-content-utils";
import {
  fetchFileContent,
  fetchMainEntityFiles,
} from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  ResourceKind,
  fetchResources,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1Resource,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, get } from "svelte/store";
import { FileArtifact } from "./file-artifact";

export class FileArtifacts {
  /**
   * Map of all files and its individual store
   */
  private readonly artifacts: Record<string, FileArtifact> = {};

  // Actions

  public async init(queryClient: QueryClient, instanceId: string) {
    const resources = await fetchResources(queryClient, instanceId);
    for (const resource of resources) {
      switch (resource.meta?.name?.kind) {
        case ResourceKind.Source:
        case ResourceKind.Connector:
        case ResourceKind.Model:
        case ResourceKind.MetricsView:
        case ResourceKind.Component:
        case ResourceKind.Dashboard:
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

    const files = await fetchMainEntityFiles(queryClient, instanceId);
    const missingFiles = files.filter(
      (f) => !this.artifacts[f] || !get(this.artifacts[f].name)?.kind,
    );

    await Promise.all(
      missingFiles.map((filePath) =>
        fetchFileContent(queryClient, instanceId, filePath).then(
          (fileContents) => {
            const artifact =
              this.artifacts[filePath] ?? new FileArtifact(filePath);
            const newName = parseKindAndNameFromFile(filePath, fileContents);
            if (newName) artifact.name.set(newName);
            this.artifacts[filePath] ??= artifact;
          },
        ),
      ),
    );
  }

  public async fileUpdated(filePath: string) {
    this.artifacts[filePath] ??= new FileArtifact(filePath);
    const fileContents = await fetchFileContent(
      queryClient,
      get(runtime).instanceId,
      filePath,
    );
    const newName = parseKindAndNameFromFile(filePath, fileContents);
    if (newName) this.artifacts[filePath].name.set(newName);
    this.artifacts[filePath].updateRemoteContent(fileContents);
  }

  /**
   * This is called when an artifact is deleted.
   */
  public fileDeleted(filePath: string) {
    // 2-way delete to handle race condition with delete from file and resource watchers
    // TODO: avoid this if `name` is undefined - event from resource will not be present
    if (this.artifacts[filePath]?.deleted) {
      // already marked for delete in resourceDeleted, delete from cache
      delete this.artifacts[filePath];
    } else if (this.artifacts[filePath]) {
      // seeing delete for the 1st time, mark for delete
      this.artifacts[filePath].deleted = true;
    }
  }

  public resourceDeleted(name: V1ResourceName) {
    const artifact = this.findFileArtifact(
      (name.kind ?? "") as ResourceKind,
      name.name ?? "",
    );
    if (!artifact) return;
    // 2-way delete to handle race condition with delete from file and resource watchers
    if (artifact.deleted) {
      // already marked for delete in fileDeleted, delete from cache
      delete this.artifacts[artifact.path];
    } else {
      // seeing delete for the 1st time, mark for delete
      artifact.deleted = true;
    }
  }

  public updateArtifacts(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateAll(resource);
    });
  }

  public updateReconciling(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateReconciling(resource);
    });
  }

  public updateLastUpdated(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateLastUpdated(resource);
    });
  }

  /**
   * This is called when a resource is deleted either because file was deleted or it errored out.
   */
  public softDeleteResource(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.artifacts[filePath]?.softDeleteResource();
    });
  }

  // Selectors

  public getFileArtifact(filePath: string) {
    this.artifacts[filePath] ??= new FileArtifact(filePath);
    return this.artifacts[filePath];
  }

  public findFileArtifact(resKind: ResourceKind, resName: string) {
    for (const filePath in this.artifacts) {
      const name = get(this.artifacts[filePath].name);
      if (name?.kind === resKind && name?.name === resName) {
        return this.artifacts[filePath];
      }
    }
    return undefined;
  }

  /**
   * Best effort list of all reconciling resources.
   */
  public getReconcilingResourceNames() {
    const artifacts = Object.values(this.artifacts);
    return derived(
      artifacts.map((a) => a.reconciling),
      (reconcilingArtifacts) => {
        const currentlyReconciling = new Array<V1ResourceName>();
        reconcilingArtifacts.forEach((reconcilingArtifact, i) => {
          if (reconcilingArtifact) {
            currentlyReconciling.push(get(artifacts[i].name) as V1ResourceName);
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
  public getNamesForKind(kind: ResourceKind): string[] {
    return Object.values(this.artifacts)
      .filter((artifact) => get(artifact.name)?.kind === kind)
      .map((artifact) => get(artifact.name)?.name ?? "");
  }

  public async saveAll() {
    await Promise.all(
      Object.entries(this.artifacts).map(([_, artifact]) =>
        artifact.saveLocalContent(),
      ),
    );
  }

  public hasUnsaved() {
    return Object.values(this.artifacts).some((artifact) =>
      get(artifact.hasUnsavedChanges),
    );
  }
}

export const fileArtifacts = new FileArtifacts();
