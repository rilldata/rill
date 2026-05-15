import {
  PossibleFileExtensions,
  PossibleZipExtensions,
} from "@rilldata/web-common/features/sources/modal/possible-file-extensions";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export const FileTooLargeError = new Error(
  "File exceeds the maximum size. Please choose a smaller file to continue.",
);

export async function uploadFile(
  client: RuntimeClient,
  file: File,
): Promise<string> {
  const formData = new FormData();
  formData.append("file", file);

  const filePath = `data/${file.name}`;

  try {
    const url = `${client.host}/v1/instances/${client.instanceId}/files/upload/-/${filePath}`;
    const headers: Record<string, string> = {};
    const jwt = client.getJwt();
    if (jwt) headers["Authorization"] = `Bearer ${jwt}`;
    const resp = await fetch(url, { method: "POST", headers, body: formData });
    if (!resp.ok) throw new Error(`Upload failed: ${resp.status}`);
    return filePath;
  } catch (err) {
    if (err.message.includes("413")) {
      throw FileTooLargeError;
    }
    throw err;
  }
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
        resolve([]);
      }, 1000);
    };
    window.addEventListener("focus", focusHandler);
    input.multiple = multiple;
    input.accept = [...PossibleFileExtensions, ...PossibleZipExtensions].join(
      ",",
    );
    input.click();
  });
}
