// Helpers for sanitizing secret fields out of connector config
export function getSecretKeysFromConnector(conn: {
  configProperties?: Array<{ key?: string | null; secret?: boolean }>;
}): string[] {
  return (
    (conn.configProperties
      ?.filter((property) => property.secret)
      .map((property) => property.key)
      .filter(Boolean) as string[]) ?? []
  );
}

export function sanitizeValuesByKeys(
  values: Record<string, unknown>,
  keys: string[],
): Record<string, unknown> {
  const cleaned: Record<string, unknown> = { ...values };
  for (const k of keys) delete cleaned[k];
  return cleaned;
}
