import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params";
import { derived, type Readable } from "svelte/store";

/**
 * Pure function that checks whether the current browser URL represents
 * the YAML-configured default explore state.
 *
 * The browser URL is pre-cleaned against rill defaults by DashboardStateSync,
 * so absent params implicitly carry rill-default values. We check:
 * 1. Forward: no non-YAML-default params in the browser URL
 * 2. Reverse: all YAML defaults that differ from rill defaults are present
 *    in the browser URL
 */
export function isViewingDefaults(
  cleanedUrlParams: URLSearchParams,
  yamlDefaultUrlParams: URLSearchParams | undefined,
  rillDefaultUrlParams: URLSearchParams | undefined,
  rawUrlParams: URLSearchParams,
): boolean {
  // If no YAML defaults configured, never "viewing defaults"
  if (!yamlDefaultUrlParams || yamlDefaultUrlParams.size === 0) return false;
  // Forward: no non-default params in browser URL
  if (cleanedUrlParams.size !== 0) return false;
  // Reverse: YAML defaults that differ from rill defaults must be in the browser URL.
  // Params absent from the browser URL implicitly have their rill-default values
  // (because DashboardStateSync cleans matching params). So if a YAML default
  // differs from the rill default (e.g. filter) but is absent from the URL,
  // the user is NOT viewing YAML defaults.
  if (!rillDefaultUrlParams) return false;
  const significantYamlDefaults = cleanUrlParams(
    yamlDefaultUrlParams,
    rillDefaultUrlParams,
  );
  return cleanUrlParams(significantYamlDefaults, rawUrlParams).size === 0;
}

/**
 * Creates a derived store wrapper around {@link isViewingDefaults}.
 */
export function createViewingDefaultsStore(
  currentCleanedUrlParams: Readable<URLSearchParams>,
  yamlDefaultUrlParams: Readable<URLSearchParams | undefined>,
  rillDefaultUrlParams: Readable<URLSearchParams | undefined>,
  rawUrlParams: Readable<URLSearchParams>,
): Readable<boolean> {
  return derived(
    [
      currentCleanedUrlParams,
      yamlDefaultUrlParams,
      rillDefaultUrlParams,
      rawUrlParams,
    ],
    ([$cleaned, $yamlDefaults, $rillDefaults, $raw]) =>
      isViewingDefaults($cleaned, $yamlDefaults, $rillDefaults, $raw),
  );
}
