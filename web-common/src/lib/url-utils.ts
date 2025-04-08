/**
 * Copies all parameters from the source URLSearchParams object to the target URLSearchParams object,
 * modifying the target object directly. Any existing parameters in the target with the same keys
 * will be overwritten.
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
