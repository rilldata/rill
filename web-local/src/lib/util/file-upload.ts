import { goto } from "$app/navigation";
import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { get } from "svelte/store";
import {
  config,
  dataModelerService,
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
import { fetchWrapper, fetchWrapperDirect } from "./fetchWrapper";

/**
 * uploadTableFiles
 * --------
 * Attempts to upload all files passed in.
 * Will return the list of files that are not valid.
 */
function uploadTableFiles(
  files,
  duplicateValidator: (name: string) => boolean,
  incrementedNameGetter: (name: string) => string
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
    validateFile(validFile, duplicateValidator, incrementedNameGetter)
  );
  return invalidFiles;
}

async function validateFile(
  file: File,
  duplicateValidator: (name: string) => boolean,
  incrementedNameGetter: (name: string) => string
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
          incrementedNameGetter(currentTableName)
        );
      } else if (userResponse == DuplicateActions.Overwrite) {
        await uploadFile(file, tableUploadURL, currentTableName);
      }
    } else {
      await uploadFile(file, tableUploadURL, currentTableName);
    }
  } catch (err) {
    console.error(err);
  }
}

async function uploadFile(file: File, url: string, tableName?: string) {
  importOverlayVisible.set(true);

  const formData = new FormData();
  formData.append("file", file);
  formData.append("instanceId", get(runtimeStore).instanceId);
  formData.append("tableName", tableName);

  try {
    const persistentTable: PersistentTableEntity = await fetchWrapperDirect(
      `${url}/-/${file.name}`,
      "POST",
      formData,
      {}
    );
    await sourceUpdated(persistentTable.tableName);
    goto(`/source/${persistentTable.id}`);
    // do not await here. it should not block importOverlayVisible being set to false
    dataModelerService.dispatch("collectTableInfo", [persistentTable.id]);
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
  duplicateValidator: (name: string) => boolean,
  incrementedNameGetter: (name: string) => string
) {
  let invalidFiles = [];
  if (filesArray) {
    invalidFiles = uploadTableFiles(
      filesArray,
      duplicateValidator,
      incrementedNameGetter
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
  duplicateValidator: (name: string) => boolean,
  incrementedNameGetter: (name: string) => string
) {
  const files = e?.dataTransfer?.files;
  if (files) {
    handleFileUploads(
      Array.from(files),
      duplicateValidator,
      incrementedNameGetter
    );
  }
}

export async function uploadFilesWithDialog(
  duplicateValidator: (name: string) => boolean,
  incrementedNameGetter: (name: string) => string
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
        incrementedNameGetter
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
