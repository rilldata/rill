import {
  extractFileExtension,
  extractFileName,
  splitFolderAndName,
} from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  ResourceKind,
  useProjectParser,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  V1ReconcileStatus,
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGetFile,
  runtimeServicePutFile,
  type V1ParseError,
  type V1Resource,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient, QueryFunction } from "@tanstack/svelte-query";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import {
  DIRECTORIES_WITHOUT_AUTOSAVE,
  FILES_WITHOUT_AUTOSAVE,
} from "../editor/config";
import { fileArtifacts } from "./file-artifacts";
import { inferResourceKind } from "./infer-resource-kind";

const UNSUPPORTED_EXTENSIONS = [".parquet", ".db", ".db.wal"];

export class FileArtifact {
  readonly path: string;
  readonly resourceName = writable<V1ResourceName | undefined>(undefined);
  readonly inferredResourceKind = writable<ResourceKind | null | undefined>(
    undefined,
  );
  readonly reconciling = writable(false);
  readonly localContent: Writable<string | null> = writable(null);
  readonly remoteContent: Writable<string | null> = writable(null);
  readonly hasUnsavedChanges = writable(false);
  readonly fileExtension: string;
  readonly ready: Promise<boolean>;
  readonly fileTypeUnsupported: boolean;
  readonly folderName: string;
  readonly fileName: string;
  readonly disableAutoSave: boolean;
  readonly autoSave: Writable<boolean>;

  private remoteCallbacks = new Set<(content: string) => void>();
  private localCallbacks = new Set<(content: string | null) => void>();

  // Last time the state of the resource `kind/name` was updated.
  // This is updated in watch-resources and is used there to avoid
  // unnecessary calls to GetResource API.
  lastStateUpdatedOn: string | undefined;

  constructor(filePath: string) {
    const [folderName, fileName] = splitFolderAndName(filePath);

    this.path = filePath;
    this.folderName = folderName;
    this.fileName = fileName;

    this.disableAutoSave =
      FILES_WITHOUT_AUTOSAVE.includes(filePath) ||
      DIRECTORIES_WITHOUT_AUTOSAVE.includes(folderName);

    if (this.disableAutoSave) {
      this.autoSave = writable(false);
    } else {
      this.autoSave = localStorageStore<boolean>(`autoSave::${filePath}`, true);
    }

    this.fileExtension = extractFileExtension(filePath);
    this.fileTypeUnsupported = UNSUPPORTED_EXTENSIONS.includes(
      this.fileExtension,
    );

    this.onRemoteContentChange((content) => {
      if (!get(this.resourceName)) {
        this.inferredResourceKind.set(inferResourceKind(this.path, content));
      }
    });
  }

  updateRemoteContent = (content: string, alert = true) => {
    this.remoteContent.set(content);
    if (alert) {
      for (const callback of this.remoteCallbacks) {
        callback(content);
      }
    }
  };

  async fetchContent(invalidate = false) {
    const instanceId = get(runtime).instanceId;
    const queryParams = {
      path: this.path,
    };
    const queryKey = getRuntimeServiceGetFileQueryKey(instanceId, queryParams);

    if (invalidate) await queryClient.invalidateQueries(queryKey);

    const queryFn: QueryFunction<
      Awaited<ReturnType<typeof runtimeServiceGetFile>>
    > = ({ signal }) => runtimeServiceGetFile(instanceId, queryParams, signal);

    const { blob } = await queryClient.fetchQuery({
      queryKey,
      queryFn,
      staleTime: Infinity,
    });

    if (blob === undefined) {
      throw new Error("Content undefined");
    }

    this.updateRemoteContent(blob, true);
  }

  updateLocalContent = (content: string | null, alert = false) => {
    const hasUnsavedChanges = get(this.hasUnsavedChanges);
    const autoSave = get(this.autoSave);

    if (content === null) {
      this.hasUnsavedChanges.set(false);
      fileArtifacts.unsavedFiles.update((files) => {
        files.delete(this.path);
        return files;
      });
    } else if (!hasUnsavedChanges && !autoSave) {
      this.hasUnsavedChanges.set(true);
      fileArtifacts.unsavedFiles.update((files) => {
        files.add(this.path);
        return files;
      });
    }

    this.localContent.set(content);

    if (alert) {
      for (const callback of this.localCallbacks) {
        callback(content);
      }
    }
  };

  onRemoteContentChange = (callback: (content: string) => void) => {
    this.remoteCallbacks.add(callback);
    return () => this.remoteCallbacks.delete(callback);
  };

  onLocalContentChange = (callback: (content: string | null) => void) => {
    this.localCallbacks.add(callback);
    return () => this.localCallbacks.delete(callback);
  };

  revert = () => {
    this.updateLocalContent(null, true);
  };

  saveLocalContent = async () => {
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

    try {
      await runtimeServicePutFile(instanceId, {
        path: this.path,
        blob: local,
      }).catch(console.error);

      this.updateLocalContent(null);
    } catch (e) {
      console.error(e);
    }
  };

  updateAll(resource: V1Resource) {
    this.updateResourceNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
  }

  updateReconciling(resource: V1Resource) {
    this.updateResourceNameIfChanged(resource);
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
  }

  updateLastUpdated(resource: V1Resource) {
    this.updateResourceNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
  }

  softDeleteResource() {
    this.reconciling.set(false);
  }

  hardDeleteResource() {
    // To avoid a workspace flicker, first infer the *intended* resource kind
    this.inferredResourceKind.set(
      inferResourceKind(this.path, get(this.remoteContent) ?? ""),
    );

    this.resourceName.set(undefined);
    this.reconciling.set(false);
    this.lastStateUpdatedOn = undefined;
  }

  getResource = (queryClient: QueryClient, instanceId: string) => {
    return derived(this.resourceName, (name, set) =>
      useResource(
        instanceId,
        name?.name as string,
        name?.kind as ResourceKind,
        undefined,
        queryClient,
      ).subscribe(set),
    ) as ReturnType<typeof useResource<V1Resource>>;
  };

  getParseError = (queryClient: QueryClient, instanceId: string) => {
    return derived(
      useProjectParser(queryClient, instanceId),
      (projectParser) => {
        return projectParser.data?.projectParser?.state?.parseErrors?.find(
          (e) => e.filePath === this.path,
        );
      },
    );
  };

  getAllErrors = (
    queryClient: QueryClient,
    instanceId: string,
  ): Readable<V1ParseError[]> => {
    const store = derived(
      [
        useProjectParser(queryClient, instanceId),
        this.getResource(queryClient, instanceId),
      ],
      ([projectParser, resource]) => {
        if (projectParser.isFetching || resource.isFetching) {
          // to avoid flicker during a re-fetch, retain the previous value
          return get(store);
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

  getHasErrors(queryClient: QueryClient, instanceId: string) {
    return derived(
      this.getAllErrors(queryClient, instanceId),
      (errors) => errors.length > 0,
    );
  }

  getEntityName() {
    return get(this.resourceName)?.name ?? extractFileName(this.path);
  }

  private updateResourceNameIfChanged(resource: V1Resource) {
    const isSubResource = !!resource.component?.spec?.definedInDashboard;
    if (isSubResource) return;
    const curName = get(this.resourceName);
    if (
      curName?.name !== resource.meta?.name?.name ||
      curName?.kind !== resource.meta?.name?.kind
    ) {
      this.resourceName.set({
        kind: resource.meta?.name?.kind,
        name: resource.meta?.name?.name,
      });
    }
  }
}
