import { goto } from "$app/navigation";
import { runtimeServiceFileUpload } from "@rilldata/web-common/runtime-client/manual-clients";
import {
  duplicateNameChecker,
  incrementedNameGetter,
} from "@rilldata/web-local/lib/util/duplicateNameUtils";
import {
  DuplicateActions,
  duplicateSourceAction,
  duplicateSourceName,
} from "../application-state-stores/application-store";
import { importOverlayVisible } from "../application-state-stores/overlay-store";
import { notifications } from "../components/notifications";
import { FILE_EXTENSION_TO_TABLE_TYPE } from "../types";
import {
  extractFileExtension,
  getTableNameFromFile,
} from "./extract-table-name";

/**
 * Uploads all valid files.
 * If any file exists, a prompt is shown to resolve the duplicates.
 * Returns table name and file paths of all uploaded files.
 * Note: actual creation of the table with the file is not done by this method.
 */
export async function* uploadTableFiles(
  files: Array<File>,
  [models, sources]: [Array<string>, Array<string>],
  instanceId: string,
  goToIfSuccessful = true
): AsyncGenerator<{ tableName: string; filePath: string }> {
  if (!files?.length) return;
  const { validFiles, invalidFiles } = filterValidFileExtensions(files);

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

    const filePath = await uploadFile(instanceId, validFile);
    // if upload failed for any reason continue
    if (filePath) {
      lastTableName = resolvedTableName;
      yield { tableName: resolvedTableName, filePath };
    }

    importOverlayVisible.set(false);
  }

  if (lastTableName && goToIfSuccessful) {
    goto(`/source/${lastTableName}`);
  }

  if (invalidFiles.length) {
    reportFileErrors(invalidFiles);
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

export async function uploadFile(
  instanceId: string,
  file: File
): Promise<string> {
  const formData = new FormData();
  formData.append("file", file);

  const filePath = `data/${file.name}`;

  try {
    await runtimeServiceFileUpload(instanceId, filePath, formData);
    return filePath;
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
      } else {
        resolve([]);
      }
    };
    const focusHandler = () => {
      window.removeEventListener("focus", focusHandler);
      setTimeout(() => {
        console.log("focus timeout");
        resolve([]);
      }, 1000);
    };
    window.addEventListener("focus", focusHandler);
    input.multiple = multiple;
    input.accept = ".csv,.tsv,.txt,.parquet";
    input.click();
  });
}
