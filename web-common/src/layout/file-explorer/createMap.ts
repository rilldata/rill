export type DirectoryOrFile = Map<string, DirectoryOrFile | null>;

export function buildMapStructure(files: string[]): DirectoryOrFile {
  const structure = new Map<string, DirectoryOrFile | null>();
  const splitFiles = files.map((f) => f.split("/").filter(Boolean));

  const sortedFiles = splitFiles.sort((a, b) => {
    const aIsFile = a.length === 1;
    const bIsFile = b.length === 1;
    return Number(aIsFile) - Number(bIsFile) || a[0].localeCompare(b[0]);
  });

  for (const parts of sortedFiles) {
    addPathToMap(structure, parts);
  }

  return structure;
}

function addPathToMap(map: DirectoryOrFile, pathParts: string[]): void {
  if (pathParts.length === 0) return;

  const [first, ...rest] = pathParts;
  if (!first) return;

  const isFile = rest.length === 0;

  if (isFile && !map.has(first)) {
    map.set(first, null);
  }

  if (!isFile) {
    if (!map.has(first)) {
      map.set(first, new Map());
    }
    const subMap = map.get(first);
    if (subMap instanceof Map) {
      addPathToMap(subMap, rest);
    }
  }
}
