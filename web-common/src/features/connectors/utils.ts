import { V1Resource } from "../../runtime-client";
import { ResourceKind } from "../entity-management/resource-selectors";

export function getConnectorNameForResource(resource: V1Resource): string {
  if (!resource.meta?.name?.kind || !resource.meta?.name?.name) return "";

  if (resource.meta.name.kind === ResourceKind.Connector.toString()) {
    return resource.meta.name.name;
  } else if (resource.meta.name.kind === ResourceKind.Source.toString()) {
    return resource?.source?.spec?.sinkConnector ?? "";
  } else if (resource.meta.name.kind === ResourceKind.Model.toString()) {
    return resource?.model?.spec?.outputConnector ?? "";
  }

  return "";
}
