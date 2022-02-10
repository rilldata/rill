export function sanitizeTableName(tableName: string): string {
    return tableName.split("/").slice(-1)[0]
        .replace(/[-.]/g, "_");
}
