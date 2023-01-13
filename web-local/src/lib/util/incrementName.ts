// sourced from https://www.steveruiz.me/posts/incrementing-name

// Will return "1" from "table_name_1"
const INCREMENT = new RegExp(/(\d+)$/);

const INCREMENT_NICE = new RegExp(/\d+(?=\)$)/);

export function displayName(name: string) {
  // uppercase first letter
  // remote underscores
  return name
    .split("_")
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .filter((word) => word !== "")
    .join(" ");
}

/**
 * Get an incremented name (e.g. new_table_2) from a name (e.g. new_table), based on an array of
 * existing names.
 *
 * @param name The name to increment.
 * @param others The array of existing names.
 * @param displayName will return a nice name for display purposes (e.g. a dashboard)
 */
export function getName(name: string, others: string[], displayName = false) {
  const set = new Set(
    displayName ? others : others.map((other) => other.toLowerCase())
  );

  let result = name;

  while (set.has(displayName ? result : result.toLowerCase())) {
    result = INCREMENT.exec(result)?.[1]
      ? result.replace(displayName ? INCREMENT_NICE : INCREMENT, (m) =>
          (+m + 1).toString()
        )
      : // show first duplicate
      displayName
      ? `${result} (1)`
      : `${result}_1`;
  }

  return result;
}
