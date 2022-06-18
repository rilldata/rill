import type { DimensionDefinition } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { EntityClient } from "$lib/http-client/EntityClient";

const EndpointMap: {
  [key in keyof DimensionDefinition]?: { endPoint: string; field: string };
} = {
  dimensionColumn: { endPoint: "updateColumn", field: "column" },
};

export class MetricsDimensionClient extends EntityClient<DimensionDefinition> {
  public static instance: MetricsDimensionClient;
  public static create() {
    this.instance = new MetricsDimensionClient(
      EndpointMap,
      (dimId: string, metricsId: string) => {
        return `/metrics/${metricsId}/dimensions${dimId ? "/" + dimId : ""}`;
      }
    );
  }
}
