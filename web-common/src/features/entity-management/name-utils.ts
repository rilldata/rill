// sourced from https://www.steveruiz.me/posts/incrementing-name

// Will return "1" from "table_name_1"
const INCREMENT = new RegExp(/(\d+)$/);

/**
 * Get an incremented name (e.g. new_table_2) from a name (e.g. new_table), based on an array of
 * existing names.
 *
 * @param name The name to increment.
 * @param others The array of existing names.
 */
export function getName(name: string, others: string[]) {
  const set = new Set(others.map((other) => other.toLowerCase()));

  let result = name;

  while (set.has(result.toLowerCase())) {
    result = INCREMENT.exec(result)?.[1]
      ? result.replace(INCREMENT, (m) => (+m + 1).toString())
      : `${result}_1`;
  }

  return result;
}

export function getNextEntityName(
  entityNames: Array<string>,
  entityName: string
): string {
  const idx = entityNames.indexOf(entityName);
  if (idx <= 0) {
    return entityNames[idx + 1];
  } else {
    return entityNames[idx - 1];
  }
}

export function isDuplicateName(
  name: string,
  fromName: string,
  names: Array<string>
) {
  if (name.toLowerCase() === fromName.toLowerCase()) return false;
  return names.findIndex((n) => n.toLowerCase() === name.toLowerCase()) >= 0;
}
