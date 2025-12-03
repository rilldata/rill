import type {
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "../entity-management/resource-selectors";
import { resourceNameToId } from "../entity-management/resource-utils";

const downstreamKindMapping = new Map<ResourceKind, Set<ResourceKind>>([
  [ResourceKind.MetricsView, new Set([ResourceKind.Explore])],
  [ResourceKind.Source, new Set([ResourceKind.Model])],
  [ResourceKind.Model, new Set([ResourceKind.MetricsView, ResourceKind.Model])],
]);

function refsKey(refs?: V1ResourceName[] | null) {
  if (!refs?.length) return undefined;

  return refs
    .map((ref) => resourceNameToId(ref))
    .filter((key): key is string => !!key)
    .sort()
    .join("|");
}

export function findLateralResources(
  resource: V1Resource | undefined,
  allResources: V1Resource[],
) {
  if (!resource) return [];

  const resourceNameKey = resourceNameToId(resource.meta?.name);
  const resourceRefsKey = refsKey(resource.meta?.refs);

  return allResources.filter((candidate) => {
    const candidateNameKey = resourceNameToId(candidate.meta?.name);
    if (resourceNameKey && candidateNameKey === resourceNameKey) return true;

    const candidateRefsKey = refsKey(candidate.meta?.refs);
    if (!candidateRefsKey || !resourceRefsKey) return false;

    return candidateRefsKey === resourceRefsKey;
  });
}

export function findUpstreamResources(
  resources: (V1Resource | undefined)[],
  allResources: V1Resource[],
) {
  if (!resources?.length) return [];

  const referenceKeys = new Set<string>();

  resources.forEach((resource) => {
    resource?.meta?.refs?.forEach((ref) => {
      const key = resourceNameToId(ref);
      if (key) referenceKeys.add(key);
    });
  });

  if (!referenceKeys.size) return [];

  return allResources.filter((candidate) => {
    const candidateKey = resourceNameToId(candidate.meta?.name);
    return candidateKey ? referenceKeys.has(candidateKey) : false;
  });
}

export function findDownstreamResources(
  selectedResource: V1Resource | undefined,
  resources: (V1Resource | undefined)[],
  allResources: V1Resource[],
) {
  const anchorResource =
    selectedResource ?? resources.find((resource) => !!resource);
  const resourceKind = anchorResource?.meta?.name?.kind as
    | ResourceKind
    | undefined;

  if (!resourceKind) return [];

  const downstreamKinds = downstreamKindMapping.get(resourceKind);
  if (!downstreamKinds?.size) return [];

  const selectedKey = resourceNameToId(selectedResource?.meta?.name);
  const resourceKeys = resources
    .map((resource) => resourceNameToId(resource?.meta?.name))
    .filter((key): key is string => !!key);
  const resourceKeySet = new Set(resourceKeys);

  return allResources.filter((candidate) => {
    const candidateKind = candidate.meta?.name?.kind as
      | ResourceKind
      | undefined;
    if (!candidateKind || !downstreamKinds.has(candidateKind)) return false;

    return (candidate.meta?.refs ?? []).some((ref) => {
      const refKey = resourceNameToId(ref);
      if (!refKey) return false;
      if (selectedKey) return refKey === selectedKey;
      return resourceKeySet.has(refKey);
    });
  });
}
