export function parseNumber(str: string, defaultValue = null) {
    return str ? Number(str) : defaultValue;
}
