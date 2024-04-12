import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
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
import { ResourceStatus } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
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

  public deleteResource() {
    this.name.set(undefined);
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
    return derived(
      [
        useProjectParser(queryClient, instanceId),
        this.getResource(queryClient, instanceId),
      ],
      ([projectParser, resource]) => {
        if (
          projectParser.isFetching ||
          projectParser.isError ||
          resource.isFetching
        ) {
          // TODO: what should the error be for failed get resource API
          return [];
        }
        return [
          ...(
            projectParser.data?.projectParser?.state?.parseErrors ?? []
          ).filter(
            (e) => e.filePath && removeLeadingSlash(e.filePath) === this.path,
          ),
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
  }

  public getHasErrors(queryClient: QueryClient, instanceId: string) {
    return derived(
      this.getAllErrors(queryClient, instanceId),
      (errors) => errors.length > 0,
    );
  }

  public getResourceStatusStore(
    queryClient: QueryClient,
    instanceId: string,
    validator?: (res: V1Resource) => boolean,
  ) {
    return derived(
      [
        this.getResource(queryClient, instanceId),
        this.getAllErrors(queryClient, instanceId),
        useProjectParser(queryClient, instanceId),
      ],
      ([resourceResp, errors, projectParserResp]) => {
        if (projectParserResp.isError) {
          return {
            status: ResourceStatus.Errored,
            error: projectParserResp.error,
          };
        }

        if (
          errors.length ||
          (resourceResp.isError && !resourceResp.isFetching) ||
          projectParserResp.isError
        ) {
          return {
            status: ResourceStatus.Errored,
            error: resourceResp.error ?? projectParserResp.error,
          };
        }

        let isBusy: boolean;
        if (validator && resourceResp.data) {
          isBusy = !validator(resourceResp.data);
        } else {
          isBusy =
            resourceResp.isFetching ||
            resourceResp.data?.meta?.reconcileStatus !==
              V1ReconcileStatus.RECONCILE_STATUS_IDLE;
        }

        return {
          status: isBusy ? ResourceStatus.Busy : ResourceStatus.Idle,
          resource: resourceResp.data,
        };
      },
    );
  }

  public getEntityName() {
    return get(this.name)?.name ?? extractFileName(this.path);
  }

  private updateNameIfChanged(resource: V1Resource) {
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
        case ResourceKind.Chart:
        case ResourceKind.Dashboard:
          this.updateArtifacts(resource);
          break;
      }
    }

    const files = await fetchMainEntityFiles(queryClient, instanceId);
    const missingFiles = files
      .map(removeLeadingSlash)
      .filter((f) => !this.artifacts[f] || !get(this.artifacts[f].name)?.kind);
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
    filePath = removeLeadingSlash(filePath);
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
    delete this.artifacts[filePath];
  }

  public updateArtifacts(resource: V1Resource) {
    resource.meta?.filePaths?.map(removeLeadingSlash).forEach((filePath) => {
      this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateAll(resource);
    });
  }

  public updateReconciling(resource: V1Resource) {
    resource.meta?.filePaths?.map(removeLeadingSlash).forEach((filePath) => {
      this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateReconciling(resource);
    });
  }

  public updateLastUpdated(resource: V1Resource) {
    resource.meta?.filePaths?.map(removeLeadingSlash).forEach((filePath) => {
      this.artifacts[filePath] ??= new FileArtifact(filePath);
      this.artifacts[filePath].updateLastUpdated(resource);
    });
  }

  public wasRenaming(resource: V1Resource) {
    const finishedRename = !resource.meta?.renamedFrom;
    return (
      resource.meta?.filePaths?.map(removeLeadingSlash).some((filePath) => {
        return this.artifacts[filePath].renaming && finishedRename;
      }) ?? false
    );
  }

  /**
   * This is called when a resource is deleted either because file was deleted or it errored out.
   */
  public deleteResource(resource: V1Resource) {
    resource.meta?.filePaths?.map(removeLeadingSlash).forEach((filePath) => {
      this.artifacts[filePath]?.deleteResource();
    });
  }

  // Selectors

  public getFileArtifact(filePath: string) {
    filePath = removeLeadingSlash(filePath);
    this.artifacts[filePath] ??= new FileArtifact(filePath);
    return this.artifacts[filePath];
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
}

export const fileArtifacts = new FileArtifacts();
