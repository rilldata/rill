/**
 * Copies all parameters from the source URLSearchParams object to the target URLSearchParams object,
 * modifying the target object directly. Any existing parameters in the target with the same keys
 * will be overwritten.
 *
 * Note: Unlike mergeParamsWithOverwrite, this function modifies the target object directly
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

/**
 * Merges two URLSearchParams objects, where parameters from sourceParamsB overwrite
 * any matching parameters in sourceParamsA, while preserving parameters in sourceParamsA
 * that don't exist in sourceParamsB.
 *
 * @param sourceParamsA - The base URLSearchParams object
 * @param sourceParamsB - The URLSearchParams object with parameters that will overwrite matching ones in sourceParamsA
 * @returns A new URLSearchParams object with the merged parameters
 */
export function mergeParamsWithOverwrite(
  sourceParamsA: URLSearchParams,
  sourceParamsB: URLSearchParams,
): URLSearchParams {
  // Create a new URLSearchParams object to avoid modifying the originals
  const mergedParams = new URLSearchParams();

  // First, copy all parameters from sourceParamsA
  sourceParamsA.forEach((value, key) => {
    mergedParams.set(key, value);
  });

  // Then, overwrite with parameters from sourceParamsB
  sourceParamsB.forEach((value, key) => {
    mergedParams.set(key, value);
  });

  return mergedParams;
}
