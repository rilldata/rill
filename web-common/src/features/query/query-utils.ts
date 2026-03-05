/** Converts a V1StructType field type code to a display string */
export function prettyPrintType(code: string | undefined): string {
  if (!code) return "UNKNOWN";
  const normalized = code.replace(/^CODE_/, "");
  return normalized.startsWith("UNKNOWN(") ? "UNKNOWN" : normalized;
}
