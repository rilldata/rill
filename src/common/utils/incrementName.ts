// sourced from https://www.steveruiz.me/posts/incrementing-name

// Will return "(1)" from "Blog Post (1)"
const INCREMENT = new RegExp(/\s\((\d+)\)$/)

// Will return "1" from "(1)"
const INCREMENT_INT = new RegExp(/\d+(?=\)$)/)

/**
 * Get an incremented name (e.g. New page (2)) from a name (e.g. New page), based on an array of
 * existing names.
 *
 * @param name The name to increment.
 * @param others The array of existing names.
 */
export function getName(name: string, others: string[]) {
  const set = new Set(others)

  let result = name

  while (set.has(result)) {
    result = INCREMENT.exec(result)?.[1]
      ? result.replace(INCREMENT_INT, (m) => (+m + 1).toString())
      : `${result} (1)`
  }

  return result
}
