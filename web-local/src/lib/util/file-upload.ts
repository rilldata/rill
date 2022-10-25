import { goto } from "$app/navigation";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import { waitForSource } from "@rilldata/web-local/lib/components/assets/sources/sourceUtils";
import { get } from "svelte/store";
import type { V1PutRepoObjectResponse } from "web-common/src/runtime-client";
import {
  config,
  dataModelerStateService,
  DuplicateActions,
  duplicateSourceAction,
  duplicateSourceName,
  runtimeStore,
} from "../application-state-stores/application-store";
import { importOverlayVisible } from "../application-state-stores/layout-store";
import notifications from "../components/notifications";
import { sourceUpdated } from "../redux-store/source/source-apis";
import { FILE_EXTENSION_TO_TABLE_TYPE } from "../types";
import {
  extractFileExtension,
  getTableNameFromFile,
} from "./extract-table-name";
import { fetchWrapperDirect } from "./fetchWrapper";

export type DuplicateValidator = (name: string) => boolean;
export type IncrementedNameGetter = (name: string) => string;
export type UploadCallback = (
  tableName: string,
  filePath: string
) => Promise<void>;

/**
 * uploadTableFiles
 * --------
 * Attempts to upload all files passed in.
 * Will return the list of files that are not valid.
 */
function uploadTableFiles(
  files,
  duplicateValidator: DuplicateValidator,
  incrementedNameGetter: IncrementedNameGetter,
  uploadCallback: UploadCallback
) {
  const invalidFiles = [];
  const validFiles = [];

  [...files].forEach((file: File) => {
    const fileExtension = extractFileExtension(file.name);
    if (fileExtension in FILE_EXTENSION_TO_TABLE_TYPE) {
      validFiles.push(file);
    } else {
      invalidFiles.push(file);
    }
  });

  validFiles.forEach((validFile) =>
    validateFile(
      validFile,
      duplicateValidator,
      incrementedNameGetter,
      uploadCallback
    )
  );
  return invalidFiles;
}

async function validateFile(
  file: File,
  duplicateValidator: DuplicateValidator,
  incrementedNameGetter: IncrementedNameGetter,
  uploadCallback: UploadCallback
) {
  const tableUploadURL = `${config.database.runtimeUrl}/v1/repos/${
    get(runtimeStore).repoId
  }/objects/file`;

  const currentTableName = getTableNameFromFile(file.name);

  try {
    const isDuplicate = duplicateValidator(currentTableName);
    if (isDuplicate) {
      const userResponse = await getResponseFromModal(currentTableName);
      if (userResponse == DuplicateActions.Cancel) {
        return;
      } else if (userResponse == DuplicateActions.KeepBoth) {
        await uploadFile(
          file,
          tableUploadURL,
          incrementedNameGetter(currentTableName),
          uploadCallback
        );
      } else if (userResponse == DuplicateActions.Overwrite) {
        await uploadFile(
          file,
          tableUploadURL,
          currentTableName,
          uploadCallback
        );
      }
    } else {
      await uploadFile(file, tableUploadURL, currentTableName, uploadCallback);
    }
  } catch (err) {
    console.error(err);
  }
}

async function uploadFile(
  file: File,
  url: string,
  tableName: string,
  uploadCallback: UploadCallback
) {
  importOverlayVisible.set(true);

  const formData = new FormData();
  formData.append("file", file);
  formData.append("instanceId", get(runtimeStore).instanceId);
  formData.append("tableName", tableName);

  try {
    // TODO: generate client and use it in component
    const resp: V1PutRepoObjectResponse = await fetchWrapperDirect(
      `${url}/-/data/${file.name}`,
      "POST",
      formData,
      {}
    );
    await uploadCallback(tableName, resp.filePath);
    const newId = await waitForSource(
      tableName,
      dataModelerStateService.getEntityStateService(
        EntityType.Table,
        StateType.Persistent
      ).store
    );
    await sourceUpdated(tableName);
    goto(`/source/${newId}`);
  } catch (err) {
    console.error(err);
  }
  importOverlayVisible.set(false);
}

function reportFileErrors(invalidFiles: File[]) {
  notifications.send({
    message: `${invalidFiles.length} file${
      invalidFiles.length !== 1 ? "s are" : " is"
    } invalid: \n${invalidFiles.map((file) => file.name).join("\n")}`,
    detail: "Only .parquet, .csv, and .tsv files are supported",
    options: {
      persisted: true,
    },
  });
}

/** Handles the uploading of the datasets. Any invalid files will be reported
 * through reportFileErrors.
 */
function handleFileUploads(
  filesArray: File[],
  duplicateValidator: DuplicateValidator,
  incrementedNameGetter: IncrementedNameGetter,
  uploadCallback: UploadCallback
) {
  let invalidFiles = [];
  if (filesArray) {
    invalidFiles = uploadTableFiles(
      filesArray,
      duplicateValidator,
      incrementedNameGetter,
      uploadCallback
    );
  }
  if (invalidFiles.length) {
    importOverlayVisible.set(false);
    reportFileErrors(invalidFiles);
  }
}

/** a drag and drop callback to kick off a source table import */
export function onSourceDrop(
  e: DragEvent,
  duplicateValidator: DuplicateValidator,
  incrementedNameGetter: IncrementedNameGetter,
  uploadCallback: UploadCallback
) {
  const files = e?.dataTransfer?.files;
  if (files) {
    handleFileUploads(
      Array.from(files),
      duplicateValidator,
      incrementedNameGetter,
      uploadCallback
    );
  }
}

export async function uploadFilesWithDialog(
  duplicateValidator: DuplicateValidator,
  incrementedNameGetter: IncrementedNameGetter,
  uploadCallback: UploadCallback
) {
  const input = document.createElement("input");
  input.multiple = true;
  input.type = "file";
  /** an event callback when a source table file is chosen manually */
  input.onchange = (e: Event) => {
    const files = (<HTMLInputElement>e.target)?.files as FileList;
    if (files) {
      handleFileUploads(
        Array.from(files),
        duplicateValidator,
        incrementedNameGetter,
        uploadCallback
      );
    }
  };
  input.click();
}

async function getResponseFromModal(
  currentTableName
): Promise<DuplicateActions> {
  duplicateSourceName.set(currentTableName);

  return new Promise((resolve) => {
    const unsub = duplicateSourceAction.subscribe((action) => {
      if (action !== DuplicateActions.None) {
        setTimeout(unsub);
        duplicateSourceAction.set(DuplicateActions.None);
        resolve(action);
      }
    });
  });
}
