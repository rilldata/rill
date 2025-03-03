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

const RETAINED_PARAMS = ["resource", "type"];

export function mergeAndRetainParams(
  baseParams: URLSearchParams,
  newParams: URLSearchParams,
  retainedParamKeys: string[] = RETAINED_PARAMS,
): URLSearchParams {
  // Create a new URLSearchParams object to avoid modifying the originals
  const mergedParams = new URLSearchParams();

  // First, copy all parameters from newParams
  newParams.forEach((value, key) => {
    mergedParams.set(key, value);
  });

  // Then, ensure we retain specific parameters from baseParams
  // even if they weren't in newParams
  baseParams.forEach((value, key) => {
    if (retainedParamKeys.includes(key) && !newParams.has(key)) {
      mergedParams.set(key, value);
    }
  });

  return mergedParams;
}
