import { HttpStreamClient } from "$lib/http-client/HttpStreamClient";
import { EntityClient } from "$lib/http-client/EntityClient";

export class MetricsDefinitionClient extends EntityClient<any> {
  public static instance: MetricsDefinitionClient;
  public static create() {
    this.instance = new MetricsDefinitionClient({}, (metricsId: string) => {
      return `/metrics${metricsId ? "/" + metricsId : ""}`;
    });
  }

  public updateMetricsDefinitionModel(metricsDefId: string, modelId: string) {
    return HttpStreamClient.instance.request(
      `/metrics/${metricsDefId}/updateModel`,
      "POST",
      { modelId }
    );
  }

  public updateMetricsDefinitionTimestamp(
    metricsDefId: string,
    timeDimension: string
  ) {
    return HttpStreamClient.instance.request(
      `/metrics/${metricsDefId}/updateTimestamp`,
      "POST",
      { timeDimension }
    );
  }
}
