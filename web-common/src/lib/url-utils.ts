/**
 * Copies all parameters from the source URLSearchParams object to the target URLSearchParams object,
 * modifying the target object directly. Any existing parameters in the target with the same keys
 * will be overwritten.
 *
 * Note: Unlike mergeAndRetainParams, this function modifies the target object directly
 * instead of creating a new URLSearchParams object.
 *
 * @param fromSearchParams - The source URLSearchParams object
 * @param toSearchParams - The target URLSearchParams object that will be modified
 */
export function copyParamsToTarget(
  fromSearchParams: URLSearchParams,
  toSearchParams: URLSearchParams,
) {
  fromSearchParams.forEach((value, key) => {
    toSearchParams.set(key, value);
  });
}

export function copyWithAdditionalArguments(
  url: URL,
  args: Record<string, string>,
  deleteArgs: Record<string, boolean> = {},
) {
  const newUrl = new URL(url);
  for (const [key, value] of Object.entries(args)) {
    newUrl.searchParams.set(key, value);
  }
  for (const key of Object.keys(deleteArgs)) {
    newUrl.searchParams.delete(key);
  }
  return newUrl;
}

export function unorderedParamsAreEqual(
  src: URLSearchParams,
  tar: URLSearchParams,
) {
  if (src.size !== tar.size) return false;
  for (const [key, value] of src) {
    if (value !== tar.get(key)) return false;
  }
  return true;
}
