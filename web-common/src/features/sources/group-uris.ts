/** Groups the provided URIs (s3, gs, and https) into bins.
 * The key of the bin is the domain or bucket name, followed by the first path segment
 * (i.e. gs://bucket-name/path/to/file.csv -> bucket-name/path).
 * The values are
 */

import type { V1CatalogEntry } from "@rilldata/web-common/runtime-client";

export interface GroupedURIObject {
  [domainOrBucketPlusOnePath: string]: GroupedURI;
}

export interface SourceURI extends V1CatalogEntry {
  uri: string;
  abbreviatedURI: string;
}

export interface GroupedURI {
  /** the index at which all grouped URIs are no longer identical */
  endingIndex: number;
  /** the domain or bucket name + on additional path segment */
  domain: string;
  /** the connector label: gs, s3, https */
  connector: string;
  leftPart: string;
  uris: SourceURI[];
}

export function groupURIs(uris: V1CatalogEntry[]): GroupedURIObject {
  /** create the grouped URIs object. */
  const groupedURIs = uris.reduce((obj, entry) => {
    const uri = entry.path;
    const [_, __, rest] = uri.trim().split(/(gs:\/\/|s3:\/\/|https:\/\/)/);

    const components = rest.split("/");
    const domain =
      components.length > 2
        ? components.slice(0, 2).join("/")
        : components.join("/");

    obj[domain] = obj[domain] ? [...obj[domain], entry] : [entry];
    return obj;
  }, {});

  /** iterate through the keys of the grouped URIs object and
   * transform such that:
   * - the protocol used appears
   */
  for (const domain in groupedURIs) {
    const domainURIs = groupedURIs[domain];
    // march to the end of the string.
    const longestURI = Math.max(...domainURIs.map(({ path }) => path.length));
    let endingIndex = 0;
    for (let i = 0; i < longestURI; i++) {
      // the moment the set of URIs no longer matches, we stop.
      if (new Set(domainURIs.map(({ path }) => path.slice(0, i))).size !== 1) {
        endingIndex = i;
        break;
      }
    }
    const identifier = domainURIs[0].path.split(
      /(gs:\/\/|s3:\/\/|https:\/\/)/
    )[1];
    groupedURIs[domain] = {
      endingIndex,
      domain,
      connector: identifier.replace("://", ""),
      leftPart: domainURIs[0].path
        .slice(0, endingIndex - 1)
        .replace(identifier, ""),
      /** return */
      uris:
        domainURIs.length > 1
          ? domainURIs.map((entry) => ({
              ...entry,
              uri: entry.path,
              name: entry.name,
              abbreviatedURI: entry.path.slice(endingIndex - 1),
            }))
          : domainURIs.map((entry) => ({
              ...entry,
              uri: entry.path,
              name: entry.name,
              abbreviatedURI: entry.path.split("/").at(-1),
            })),
    };
  }
  return groupedURIs;
}
