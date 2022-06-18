import type { MeasureDefinition } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { EntityClient } from "$lib/http-client/EntityClient";

const EndpointMap: {
  [key in keyof MeasureDefinition]?: { endPoint: string; field: string };
} = {
  expression: { endPoint: "updateExpression", field: "expression" },
  sqlName: { endPoint: "updateSqlName", field: "sqlName" },
};

export class MetricsMeasureClient extends EntityClient<MeasureDefinition> {
  public static instance: MetricsMeasureClient;
  public static create() {
    this.instance = new MetricsMeasureClient(
      EndpointMap,
      (measureId: string, metricsId: string) => {
        return `/metrics/${metricsId}/measures${
          measureId ? "/" + measureId : ""
        }`;
      }
    );
  }
}
