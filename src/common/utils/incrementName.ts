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
  const set = new Set(others);

  let result = name;

  while (set.has(result)) {
    result = INCREMENT.exec(result)?.[1]
      ? result.replace(INCREMENT, (m) => (+m + 1).toString())
      : `${result}_1`;
  }

  return result;
}
