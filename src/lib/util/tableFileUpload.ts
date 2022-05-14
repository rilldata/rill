import {extractFileExtension} from "$lib/util/extract-table-name";
import {FILE_EXTENSION_TO_TABLE_TYPE} from "$lib/types";

/**
 * uploadTableFiles
 * --------
 * Attempts to upload all files passed in.
 * Will return the list of files that are not valid.
 */
export function uploadTableFiles(files, apiBase: string) {
    const invalidFiles = [];
    const validFiles = [];

    [...files].forEach((file:File) => {
        const fileExtension = extractFileExtension(file.name);
        if (fileExtension in FILE_EXTENSION_TO_TABLE_TYPE) {
            validFiles.push(file);
        } else {
            invalidFiles.push(file);
        }
    })

    validFiles.forEach(validFile => uploadFile(validFile, `${apiBase}/table-upload`));
    return invalidFiles;
}

export function uploadFile(file: File, url: string) {
    const formData = new FormData();
    formData.append("file", file);

    fetch(url, {
        method: "POST",
        body: formData
    })
        .then((...args) => console.error(...args))
        .catch((...args) => console.error(...args));
}
