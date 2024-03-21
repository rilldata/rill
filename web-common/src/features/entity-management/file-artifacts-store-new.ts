import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
import {
  type ResourceKind,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { MetricsConfig } from "@rilldata/web-common/features/metrics-views/metrics-internal-store";
import { getSubStore } from "@rilldata/web-common/lib/getSubStore";
import {
  V1ReconcileStatus,
  type V1Resource,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { derived, get, type Readable, writable } from "svelte/store";

export class FileArtifact {
  public readonly path = writable<string>("");

  public readonly name = writable<V1ResourceName | undefined>(undefined);

  public readonly jsonRepresentation = writable<
    MetricsConfig | Record<string, never> | undefined
  >(undefined);

  /**
   * Last time the state of the resource `kind/name` was updated
   * used to make sure we do not have unnecessary refreshes
   */
  public lastStateUpdatedOn: string | undefined;

  public constructor(filePath: string) {
    this.path.set(filePath);
  }

  public updateResource(resource: V1Resource) {
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
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
  }

  public deleteResource() {
    this.name.set(undefined);
  }
}

export class FileArtifactsStore {
  /**
   * Map of all files and whether it is reconciling or not.
   * If an entry is present here then there should be on in {@link artifacts} as well
   */
  public readonly files = writable<Record<string, boolean>>({});

  /**
   * Map of all files and its individual store
   */
  private readonly artifacts: Record<string, FileArtifact> = {};

  // Actions

  public updateFile(filePath: string) {
    if (get(this.files)[filePath]) return;
    this.artifacts[filePath] = new FileArtifact(filePath);
    this.files.update((f) => {
      f[filePath] = false;
      return f;
    });
  }

  public deleteFile(filePath: string) {
    delete this.artifacts[filePath];
    this.files.update((f) => {
      delete f[filePath];
      return f;
    });
  }

  public setResource(resource: V1Resource) {
    this.files.update((f) => {
      resource.meta?.filePaths?.forEach((filePath) => {
        if (!(filePath in f)) return;
        f[filePath] =
          resource.meta?.reconcileStatus ===
          V1ReconcileStatus.RECONCILE_STATUS_RUNNING;
      });
      return f;
    });
    resource.meta?.filePaths?.forEach((filePath) => {
      this.artifacts[filePath]?.updateResource(resource);
    });
  }

  public deleteResource(resource: V1Resource) {
    resource.meta?.filePaths?.forEach((filePath) => {
      this.artifacts[filePath]?.deleteResource();
    });
  }

  // Selectors

  public getFileArtifact(filePath: string) {
    return getSubStore(
      this.files,
      this.artifacts,
      filePath,
      // Dummy store to not break component code
      new FileArtifact(filePath),
    );
  }

  public getLastStateUpdatedOn(filePath: string) {
    return this.artifacts[filePath]?.lastStateUpdatedOn;
  }

  public getResourceNameForFile(filePath: string) {
    return derived(this.getFileArtifact(filePath), (artifact, set) =>
      derived(artifact.name, (resourceName) => resourceName).subscribe(set),
    ) as Readable<V1ResourceName | undefined>;
  }

  public getReconcilingItems() {
    return derived(this.files, (files) => {
      const currentlyReconciling = new Array<V1ResourceName>();
      for (const filePath in files) {
        const name = get(this.artifacts[filePath]?.name);
        if (files[filePath] && name) {
          currentlyReconciling.push(name);
        }
      }
      return currentlyReconciling;
    });
  }

  // Complex selectors based on resource API

  public getResourceForFile(
    queryClient: QueryClient,
    instanceId: string,
    filePath: string,
  ) {
    return derived(this.getResourceNameForFile(filePath), (resourceName, set) =>
      useResource(
        instanceId,
        resourceName?.name as string,
        resourceName?.kind as ResourceKind,
        undefined,
        queryClient,
      ).subscribe(set),
    ) as ReturnType<typeof useResource<V1Resource>>;
  }

  public getAllErrorsForFile(
    queryClient: QueryClient,
    instanceId: string,
    filePath: string,
  ) {
    return derived(
      [
        useProjectParser(queryClient, instanceId),
        this.getResourceForFile(queryClient, instanceId, filePath),
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
            (e) => e.filePath && removeLeadingSlash(e.filePath) === filePath,
          ),
          ...(resource.data?.meta?.reconcileError
            ? [
                {
                  filePath,
                  message: resource.data.meta.reconcileError,
                },
              ]
            : []),
        ];
      },
      [],
    );
  }

  public getFileHasErrors(
    queryClient: QueryClient,
    instanceId: string,
    filePath: string,
  ) {
    return derived(
      this.getAllErrorsForFile(queryClient, instanceId, filePath),
      (errors) => errors.length > 0,
    );
  }
}

export const newFileArtifactStore = new FileArtifactsStore();
