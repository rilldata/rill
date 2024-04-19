import { page } from "$app/stores";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
  ResourceKindToScreenMap,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import { get } from "svelte/store";

export function emitNavigationEventFromKindAndName(
  kind: ResourceKind,
  name: string,
) {
  return behaviourEvent.fireNavigationEvent(
    name,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
    getScreenNameFromPage(),
    ResourceKindToScreenMap[kind] ?? MetricsEventScreenName.Unknown,
  );
}

export function getScreenNameFromPage() {
  const file = get(page).params.file;
  if (!file) return MetricsEventScreenName.Unknown;
  const fileArtifact = fileArtifacts.getFileArtifact(file);
  const resName = get(fileArtifact.name);
  if (!resName?.kind) return MetricsEventScreenName.Unknown;
  return (
    (ResourceKindToScreenMap[resName.kind] as MetricsEventScreenName) ??
    MetricsEventScreenName.Unknown
  );
}
