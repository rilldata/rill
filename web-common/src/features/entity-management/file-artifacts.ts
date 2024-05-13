import { parseKindAndNameFromFile } from "@rilldata/web-common/features/entity-management/file-content-utils";
import {
  fetchFileContent,
  fetchMainEntityFiles,
} from "@rilldata/web-common/features/entity-management/file-selectors";
import {
  fetchResources,
  ResourceKind,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetResourceQueryKey,
  type V1ParseError,
  V1ReconcileStatus,
  type V1Resource,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, get, type Readable, writable } from "svelte/store";

export class FileArtifact {
  public readonly path: string;

  public readonly name = writable<V1ResourceName | undefined>(undefined);

  public readonly reconciling = writable<boolean>(false);

  /**
   * Used to check if a file has finished renaming.
   *
   * Reconciler uses meta.renamedFrom internally to track it.
   * It is unset once rename is complete.
   */
  public renaming = false;

  /**
   * Last time the state of the resource `kind/name` was updated.
   * This is updated in watch-resources and is used there to avoid unnecessary calls to GetResource API.
   */
  public lastStateUpdatedOn: string | undefined;

  public hasTable = false;

  public deleted = false;

  public constructor(filePath: string) {
    this.path = filePath;
  }

  // actions

  public updateAll(resource: V1Resource) {
    this.updateNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
    this.renaming = !!resource.meta?.renamedFrom;
    this.hasTable = resourceHasTable(resource);
    this.deleted = false;
  }

  public updateReconciling(resource: V1Resource) {
    this.updateNameIfChanged(resource);
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
  }

  public updateLastUpdated(resource: V1Resource) {
    this.updateNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
  }

  public softDeleteResource() {
    this.reconciling.set(false);
  }

  // selectors

  public getResource(queryClient: QueryClient, instanceId: string) {
    return derived(this.name, (name, set) =>
      useResource(
        instanceId,
        name?.name as string,
        name?.kind as ResourceKind,
        undefined,
        queryClient,
      ).subscribe(set),
    ) as ReturnType<typeof useResource<V1Resource>>;
  }

  public getAllErrors(
    queryClient: QueryClient,
    instanceId: string,
  ): Readable<V1ParseError[]> {
    const store = derived(
      [
        this.name,
        useProjectParser(queryClient, instanceId),
        this.getResource(queryClient, instanceId),
      ],
      ([name, projectParser, resource]) => {
        if (
          projectParser.isFetching ||
          resource.isFetching ||
          // retain old state while reconciling
          resource.data?.meta?.reconcileStatus ===
            V1ReconcileStatus.RECONCILE_STATUS_RUNNING ||
          resource.data?.meta?.reconcileStatus ===
            V1ReconcileStatus.RECONCILE_STATUS_PENDING
        ) {
          // to avoid flicker during a re-fetch, retain the previous value
          return get(store);
        }
        if (
          // not having a name will signify a non-entity file
          !name?.kind ||
          projectParser.isError
        ) {
          return [];
        }
        return [
          ...(
            projectParser.data?.projectParser?.state?.parseErrors ?? []
          ).filter((e) => e.filePath === this.path),
          ...(resource.data?.meta?.reconcileError
            ? [
                {
                  filePath: this.path,
                  message: resource.data.meta.reconcileError,
                },
              ]
            : []),
        ];
      },
      [],
    );
    return store;
  }

  public getHasErrors(queryClient: QueryClient, instanceId: string) {
    return derived(
      this.getAllErrors(queryClient, instanceId),
      (errors) => errors.length > 0,
    );
  }

  public getEntityName() {
    return get(this.name)?.name ?? extractFileName(this.path);
  }

  private updateNameIfChanged(resource: V1Resource) {
    const isSubResource = !!resource.component?.spec?.definedInDashboard;
    if (isSubResource) return;
    const curName = get(this.name);
    if (
      curName?.name !== resource.meta?.name?.name ||
      curName?.kind !== resource.meta?.name?.kind
    ) {
      this.name.set({
        kind: resource.meta?.name?.kind,
        name: resource.meta?.name?.name,
      });
    }
  }
}

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
    if (this.artifacts[filePath] && get(this.artifacts[filePath].name)?.kind)
      return;

    this.artifacts[filePath] ??= new FileArtifact(filePath);
    const fileContents = await fetchFileContent(
      queryClient,
      get(runtime).instanceId,
      filePath,
    );
    const newName = parseKindAndNameFromFile(filePath, fileContents);
    if (newName) this.artifacts[filePath].name.set(newName);
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

  public tableStatusChanged(resource: V1Resource) {
    const hadTable =
      resource.meta?.filePaths?.some((filePath) => {
        return this.artifacts[filePath].hasTable;
      }) ?? false;
    const hasTable = resourceHasTable(resource);
    return hadTable !== hasTable;
  }

  public wasRenaming(resource: V1Resource) {
    const finishedRename = !resource.meta?.renamedFrom;
    return (
      resource.meta?.filePaths?.some((filePath) => {
        return this.artifacts[filePath].renaming && finishedRename;
      }) ?? false
    );
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
}

function resourceHasTable(resource: V1Resource) {
  return (
    (!!resource.model && !!resource.model.state?.table) ||
    (!!resource.source && !!resource.source.state?.table)
  );
}

export const fileArtifacts = new FileArtifacts();
