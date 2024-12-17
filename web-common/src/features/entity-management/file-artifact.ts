import {
  extractFileExtension,
  splitFolderAndFileName,
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
  FILE_SAVE_DEBOUNCE_TIME,
  FILES_WITHOUT_AUTOSAVE,
} from "../editor/config";
import { fileArtifacts } from "./file-artifacts";
import { inferResourceKind } from "./infer-resource-kind";
import { debounce } from "@rilldata/web-common/lib/create-debouncer";
import { AsyncSaveState } from "./async-save-state";

const UNSUPPORTED_EXTENSIONS = [
  // Data formats
  ".db",
  ".db.wal",
  ".parquet",
  ".xls",
  ".xlsx",

  // Image formats
  ".png",
  ".jpg",
  ".jpeg",
  ".gif",
  ".svg",

  // Document formats
  ".pdf",
  ".doc",
  ".docx",
  ".ppt",
  ".pptx",
];

export class FileArtifact {
  readonly path: string;
  readonly resourceName = writable<V1ResourceName | undefined>(undefined);
  readonly inferredResourceKind = writable<ResourceKind | null | undefined>(
    undefined,
  );
  readonly reconciling = writable(false);
  readonly merging = writable(false);
  readonly localContent: Writable<string | null> = writable(null);
  readonly remoteContent: Writable<string | null> = writable(null);
  readonly inConflict = writable(false);
  readonly hasUnsavedChanges = derived(
    this.localContent,
    ($local) => $local !== null,
  );
  readonly saveState = new AsyncSaveState();
  readonly saveEnabled = derived(
    [this.saveState.saving, this.hasUnsavedChanges],
    ([saving, hasUnsavedChanges]) => !saving && hasUnsavedChanges,
  );

  readonly fileExtension: string;
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
    const [folderName, fileName] = splitFolderAndFileName(filePath);

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
      const inferred = inferResourceKind(filePath, content);

      if (inferred) this.inferredResourceKind.set(inferred);
    });
  }

  fetchContent = async (invalidate = false) => {
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

    this.updateRemoteContent(blob);
    return blob;
  };

  private updateRemoteContent = (newRemoteContent: string) => {
    const currentRemoteContent = get(this.remoteContent);
    const localContent = get(this.localContent);
    const remoteContentHasChanged = currentRemoteContent !== newRemoteContent;

    if (!remoteContentHasChanged) return;

    this.remoteContent.set(newRemoteContent);
    this.saveState.resolve();

    if (localContent !== null) {
      if (localContent === newRemoteContent) {
        // This is the primary sequence that happens after saving
        // Local content is not null, user initiates a save, we receive a FILE_WRITE event
        // After fetching the remote content, it should match the local content
        // So, we revert our local content store
        this.revert();
      } else {
        // This is the secondary sequence wherein a file is saved in an external editor
        // We receive a FILE_EVENT_WRITE event and the remote content is different from local content
        // This can also happen when a file is savedmand then edited
        // in the application before the event is received
        // In this case, we ignore updates that happen within 2 seconds of the last save
        // If we receive an update after 2 seconds, we can "safely" assume
        // that the user has made conflicting changes externally
        if (Date.now() - this.saveState.lastSaveTime > 2000) {
          this.inConflict.set(true);
        }
      }
    }

    for (const callback of this.remoteCallbacks) {
      callback(newRemoteContent);
    }
  };

  updateLocalContent = (
    newContent: string | null,
    alert = false,
    autoSave = get(this.autoSave),
  ) => {
    const currentLocalContent = get(this.localContent);

    if (currentLocalContent === newContent) return;

    this.localContent.set(newContent);

    if (autoSave && newContent !== null) {
      this.debounceSave(newContent);
    } else if (newContent === null) {
      fileArtifacts.unsavedFiles.delete(this.path);
    } else {
      fileArtifacts.unsavedFiles.add(this.path);
    }

    if (alert) {
      for (const callback of this.localCallbacks) {
        callback(newContent);
      }
    }
  };

  saveLocalContent = async () => {
    const saveEnabled = get(this.saveEnabled);
    if (!saveEnabled) return;
    await this.saveContent(get(this.localContent) ?? "");
  };

  saveContent = async (blob: string) => {
    const instanceId = get(runtime).instanceId;

    // Optimistically update the query
    queryClient.setQueryData(
      getRuntimeServiceGetFileQueryKey(instanceId, {
        path: this.path,
      }),
      {
        blob,
      },
    );

    try {
      const fileSavePromise = this.saveState.initiateSave();

      await runtimeServicePutFile(instanceId, {
        path: this.path,
        blob,
      });

      await fileSavePromise;
    } catch {
      this.saveState.reject(new Error("Unable to save file."));
    }
  };

  debounceSave = debounce(this.saveContent, FILE_SAVE_DEBOUNCE_TIME);

  onRemoteContentChange = (
    callback: (content: string, force?: true) => void,
  ) => {
    this.remoteCallbacks.add(callback);
    return () => this.remoteCallbacks.delete(callback);
  };

  onLocalContentChange = (callback: (content: string | null) => void) => {
    this.localCallbacks.add(callback);
    return () => this.localCallbacks.delete(callback);
  };

  revert = (alert = false) => {
    this.merging.set(false);
    this.inConflict.set(false);

    this.updateLocalContent(null, alert);
  };

  updateResource(resource: V1Resource) {
    this.updateResourceNameIfChanged(resource);
    this.lastStateUpdatedOn = resource.meta?.stateUpdatedOn;
    this.reconciling.set(
      resource.meta?.reconcileStatus ===
        V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
    );
  }

  hardDeleteResource() {
    // To avoid a workspace flicker, first infer the *intended* resource kind
    const inferred = inferResourceKind(
      this.path,
      get(this.remoteContent) ?? "",
    );

    const curName = get(this.resourceName);
    if (inferred) {
      this.inferredResourceKind.set(inferred);
    } else if (curName && curName.kind) {
      this.inferredResourceKind.set(curName.kind as ResourceKind);
    }

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
        {
          queryClient,
        },
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

  private updateResourceNameIfChanged(resource: V1Resource) {
    const isSubResource = !!resource.component?.spec?.definedInCanvas;
    if (isSubResource) return;

    const curName = get(this.resourceName);

    // Much code currently assumes that a file is associated with 0 or 1 resource.
    // However, files for legacy Metrics Views generate 2 resources: a Metrics View and an Explore.
    // HACK: for files for legacy Metrics Views, ignore the Explore resource.
    if (
      curName?.kind === ResourceKind.MetricsView &&
      resource.meta?.name?.kind === ResourceKind.Explore
    ) {
      return;
    }

    const didResourceNameChange =
      curName?.name !== resource.meta?.name?.name ||
      curName?.kind !== resource.meta?.name?.kind;

    if (didResourceNameChange) {
      this.resourceName.set({
        kind: resource.meta?.name?.kind,
        name: resource.meta?.name?.name,
      });
    }
  }
}
