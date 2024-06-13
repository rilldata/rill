import {
  extractFileName,
  extractFileExtension,
} from "@rilldata/web-common/features/entity-management/file-path-utils";

import {
  ResourceKind,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  V1ReconcileStatus,
  type V1ParseError,
  type V1Resource,
  type V1ResourceName,
  getRuntimeServiceGetFileQueryKey,
  V1GetFileResponse,
  runtimeServiceGetFile,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryObserverResult } from "@tanstack/svelte-query";
import {
  derived,
  get,
  type Readable,
  writable,
  type Writable,
} from "svelte/store";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";

const UNSUPPORTED_EXTENSIONS = [".parquet", ".db", ".db.wal"];

export class FileArtifact {
  public readonly path: string;

  public readonly name = writable<V1ResourceName | undefined>(undefined);

  public readonly reconciling = writable<boolean>(false);

  /**
   * Last time the state of the resource `kind/name` was updated.
   * This is updated in watch-resources and is used there to avoid unnecessary calls to GetResource API.
   */
  public lastStateUpdatedOn: string | undefined;

  public deleted = false;

  public localContent: Writable<string | null> = writable(null);
  public remoteContent: Writable<string | null> = writable(null);
  public hasUnsavedChanges = derived(this.localContent, (localContent) => {
    return localContent !== null;
  });
  public fileExtension: string;
  public fileQuery: Readable<QueryObserverResult<V1GetFileResponse, HTTPError>>;
  public ready: Promise<boolean>;
  private remoteCallbacks = new Set<(content: string) => void>();
  private localCallbacks = new Set<(content: string | null) => void>();

  constructor(filePath: string) {
    this.path = filePath;

    this.fileExtension = extractFileExtension(filePath);
    const fileTypeUnsupported = UNSUPPORTED_EXTENSIONS.includes(
      this.fileExtension,
    );

    const queryKey = getRuntimeServiceGetFileQueryKey(get(runtime).instanceId, {
      path: filePath,
    });

    // Initial data fetch
    this.ready = fileTypeUnsupported
      ? Promise.resolve(false)
      : queryClient
          .fetchQuery({
            queryKey,
            queryFn: () =>
              runtimeServiceGetFile(get(runtime).instanceId, {
                path: filePath,
              }),
          })
          .then(({ blob }) => {
            if (blob === undefined) return false;
            this.updateRemoteContent(blob, false);
            return true;
          })
          .catch((e) => {
            console.error(e);
            return false;
          });
  }

  public updateRemoteContent = (content: string, alert = true) => {
    this.remoteContent.set(content);
    if (alert) {
      for (const callback of this.remoteCallbacks) {
        callback(content);
      }
    }
  };

  public updateLocalContent = (content: string | null, alert = false) => {
    this.localContent.set(content);
    if (alert) {
      for (const callback of this.localCallbacks) {
        callback(content);
      }
    }
  };

  public onRemoteContentChange = (callback: (content: string) => void) => {
    this.remoteCallbacks.add(callback);
    return () => this.remoteCallbacks.delete(callback);
  };

  public onLocalContentChange = (
    callback: (content: string | null) => void,
  ) => {
    this.localCallbacks.add(callback);
    return () => this.localCallbacks.delete(callback);
  };

  public revert = () => {
    this.updateLocalContent(null, true);
  };

  public saveLocalContent = async () => {
    const local = get(this.localContent);
    if (local === null) return;

    const blob = get(this.localContent);

    const instanceId = get(runtime).instanceId;
    const key = getRuntimeServiceGetFileQueryKey(instanceId, {
      path: this.path,
    });

    queryClient.setQueryData(key, {
      blob,
    });

    await runtimeServicePutFile(instanceId, {
      path: this.path,
      blob: local,
    }).catch(console.error);

    this.localContent.set(null);
  };

  public updateAll(resource: V1Resource) {
    this.updateNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
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

  public getAllErrors = (
    queryClient: QueryClient,
    instanceId: string,
  ): Readable<V1ParseError[]> => {
    const store = derived(
      [
        this.name,
        useProjectParser(queryClient, instanceId),
        this.getResource(queryClient, instanceId),
      ],
      ([name, projectParser, resource]) => {
        if (projectParser.isFetching || resource.isFetching) {
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
  };

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
