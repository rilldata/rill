const FILE_PATH_SPLIT_REGEX = /\//;
const FILE_EXTENSION_REGEX = /\.(.*)$/;
export const INVALID_CHARS = /[^a-zA-Z_\d]/g;

export function extractTableName(filePath: string): string {
    const fileName = filePath.split(FILE_PATH_SPLIT_REGEX).slice(-1)[0];
    return fileName.replace(FILE_EXTENSION_REGEX, "");
}

export function extractFileExtension(filePath: string): string {
    const fileExtensionMatch = FILE_EXTENSION_REGEX.exec(filePath);
    return fileExtensionMatch?.[1];
}

export function sanitizeTableName(tableName: string): string {
    return tableName.replace(INVALID_CHARS, "_");
}
