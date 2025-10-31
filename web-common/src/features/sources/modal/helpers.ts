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
