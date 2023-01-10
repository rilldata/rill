import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";
import type { FileArtifactsData } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";

export interface Reference {
  reference: string;
  type: "from" | "join";
  index: number;
  referenceIndex: number;
}

/**
 *
 * @param sql the sql query
 * @returns {Reference[]} an array of reference locations
 */
// TODO: add caching to this method
export function getTableReferences(sql: string): Array<Reference> {
  if (!sql) return [];
  // eslint-disable-next-line no-useless-escape
  const regex = /(?:from|join)\s+([a-zA-z0-9_.]+|"[a-zA-z0-9\.\-_\/:\s~]+")/gim;
  return [...sql.matchAll(regex)].map((match) => {
    return {
      reference: match[1],
      type: match[0].split(/\s/)[0].toLowerCase() as "from" | "join",
      index: match.index,
      referenceIndex: match.index + match[0].length - match[1].length,
    };
  });
}

const ProtocolMatcher = /^(?:https?|s3|gs|file):\/\//;

export function getEmbeddedReferences(sql: string): Array<Reference> {
  const dedupe = new Set<string>();
  const references = getTableReferences(sql);
  const embeddedSources = new Array<Reference>();
  for (const reference of references) {
    if (dedupe.has(reference.reference)) continue;
    const tableRef = reference.reference.substring(
      1,
      reference.reference.length - 1
    );
    dedupe.add(tableRef);
    if (!tableRef.match(/\//) && !ProtocolMatcher.test(tableRef)) continue;

    embeddedSources.push(reference);
  }

  return embeddedSources;
}

export function getTableName(
  ref: Reference,
  entities: Record<string, FileArtifactsData>
): string {
  if (!ref.reference) return "";
  const tableRef = ref.reference.substring(1, ref.reference.length - 1);
  if (!tableRef.match(/\//) && !ProtocolMatcher.test(tableRef))
    return ref.reference;
  return entities[tableRef]?.name ?? "";
}

function sourcePathMatchesReference(source: V1CatalogEntry, table: Reference) {
  return (
    `"${source.path}"` === table.reference ||
    `'${source.path}'` === table.reference
  );
}

function sourceTableReferenceIsEmbedded(
  table: Reference,
  embeddedSources: V1CatalogEntry[]
) {
  // check to see if the quoted version of this reference is in the embedded sources
  return embeddedSources?.some((source) =>
    sourcePathMatchesReference(source, table)
  );
}

/**
 * Returns the name of the table reference, or the cached name of the embedded source.
 */
export function getMatchingCatalogReference(
  table: Reference,
  embeddedSources: V1CatalogEntry[],
  existingEntities: Record<string, FileArtifactsData>
) {
  // if this reference is embedded, return the cached name
  if (sourceTableReferenceIsEmbedded(table, embeddedSources)) {
    return embeddedSources?.find((source) =>
      sourcePathMatchesReference(source, table)
    ).name;
  } else {
    return getTableName(table, existingEntities);
  }
}
