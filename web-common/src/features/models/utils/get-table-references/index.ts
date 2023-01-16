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
export function getTableReferences(sql: string) {
  if (!sql) return [];
  // eslint-disable-next-line no-useless-escape
  const regex = /(?:from|join)\s+([a-zA-z0-9_.]+|"[a-zA-z0-9\.\-_\/:\s]+")/gim;
  return [...sql.matchAll(regex)].map((match) => {
    return {
      reference: match[1],
      type: match[0].split(/\s/)[0].toLowerCase(),
      index: match.index,
      referenceIndex: match.index + match[0].length - match[1].length,
    };
  });
}
