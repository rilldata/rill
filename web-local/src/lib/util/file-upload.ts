import { goto } from "$app/navigation";
import {
  EntityType,
  StateType,
} from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { PersistentModelEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { PersistentTableEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import {
  duplicateNameChecker,
  incrementedNameGetter,
} from "@rilldata/web-local/common/utils/duplicateNameUtils";
import { waitForSource } from "@rilldata/web-local/lib/components/assets/sources/sourceUtils";
import type { V1PutRepoObjectResponse } from "web-common/src/runtime-client";
import {
  config,
  dataModelerStateService,
  DuplicateActions,
  duplicateSourceAction,
  duplicateSourceName,
  RuntimeState,
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

/**
 * Uploads all valid files.
 * If any file exists, a prompt is shown to resolve the duplicates.
 * Returns table name and file paths of all uploaded files.
 * Note: actual creation of the table with the file is not done by this method.
 */
export async function* uploadTableFiles(
  files: Array<File>,
  [models, sources]: [
    Array<PersistentModelEntity>,
    Array<PersistentTableEntity>
  ],
  runtimeState: RuntimeState
): AsyncGenerator<{ tableName: string; filePath: string }> {
  if (!files) return;
  const { validFiles, invalidFiles } = filterValidFileExtensions(files);

  if (invalidFiles.length) {
    reportFileErrors(invalidFiles);
  }

  const tableUploadURL = `${config.database.runtimeUrl}/v1/repos/${runtimeState.repoId}/objects/file`;
  let lastTableName: string;

  for (const validFile of validFiles) {
    // check if the file is already present. get the file and
    const resolvedTableName = await checkForDuplicate(
      validFile,
      (name) => duplicateNameChecker(name, models, sources),
      (name) => incrementedNameGetter(name, models, sources)
    );
    // if there was a duplicate and cancel was clicked then we do not upload
    if (!resolvedTableName) continue;

    importOverlayVisible.set(true);

    const filePath = await uploadFile(tableUploadURL, validFile);
    // if upload failed for any reason continue
    if (filePath) {
      lastTableName = resolvedTableName;
      yield { tableName: resolvedTableName, filePath };
      await sourceUpdated(resolvedTableName);
    }

    importOverlayVisible.set(false);
  }

  if (lastTableName) {
    const newId = await waitForSource(
      lastTableName,
      dataModelerStateService.getEntityStateService(
        EntityType.Table,
        StateType.Persistent
      ).store
    );
    goto(`/source/${newId}`);
  }
}

function filterValidFileExtensions(files: Array<File>): {
  validFiles: Array<File>;
  invalidFiles: Array<File>;
} {
  const validFiles = [];
  const invalidFiles = [];

  files.forEach((file: File) => {
    const fileExtension = extractFileExtension(file.name);
    if (fileExtension in FILE_EXTENSION_TO_TABLE_TYPE) {
      validFiles.push(file);
    } else {
      invalidFiles.push(file);
    }
  });

  return { validFiles, invalidFiles };
}

/**
 * Checks if the file already exists.
 * If it does then prompt the user on what to do.
 * Return next available name with a number appended if user decides to keep both.
 * Return the table name extracted from file name in all other cases.
 */
async function checkForDuplicate(
  file: File,
  duplicateValidator: (name: string) => boolean,
  incrementedNameGetter: (name: string) => string
): Promise<string> {
  const currentTableName = getTableNameFromFile(file.name);

  try {
    const isDuplicate = duplicateValidator(currentTableName);
    if (isDuplicate) {
      const userResponse = await getResponseFromModal(currentTableName);
      if (userResponse == DuplicateActions.Cancel) {
        return;
      } else if (userResponse == DuplicateActions.KeepBoth) {
        return incrementedNameGetter(currentTableName);
      } else if (userResponse == DuplicateActions.Overwrite) {
        return currentTableName;
      }
    } else {
      return currentTableName;
    }
  } catch (err) {
    console.error(err);
  }

  return undefined;
}

export async function uploadFile(url: string, file: File): Promise<string> {
  const formData = new FormData();
  formData.append("file", file);

  try {
    // TODO: generate client and use it in component
    const resp: V1PutRepoObjectResponse = await fetchWrapperDirect(
      `${url}/-/data/${file.name}`,
      "POST",
      formData,
      {}
    );
    return resp.file_path;
  } catch (err) {
    console.error(err);
  }

  return undefined;
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

export function openFileUploadDialog(multiple = true) {
  return new Promise<Array<File>>((resolve) => {
    const input = document.createElement("input");
    input.multiple = true;
    input.type = "file";
    /** an event callback when a source table file is chosen manually */
    input.onchange = (e: Event) => {
      const files = (<HTMLInputElement>e.target)?.files as FileList;
      if (files) {
        resolve(Array.from(files));
      }
    };
    input.multiple = multiple;
    input.click();
  });
}
