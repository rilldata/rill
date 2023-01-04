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
