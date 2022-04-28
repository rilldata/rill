import {extractFileExtension} from "$lib/util/extract-table-name";
import {FILE_EXTENSION_TO_TABLE_TYPE} from "$lib/types";

export function uploadTableFiles(files, apiBase: string) {
    const validFiles = [...files].filter((file: File) => {
        const fileExtension = extractFileExtension(file.name);
        return fileExtension in FILE_EXTENSION_TO_TABLE_TYPE;
    });
    validFiles.forEach(validFile => uploadFile(validFile, `${apiBase}/table-upload`));
}

export function uploadFile(file: File, url: string) {
    const formData = new FormData();
    formData.append("file", file);

    fetch(url, {
        method: "POST",
        body: formData
    })
        .then((...args) => console.log(...args))
        .catch((...args) => console.log(...args));
}
