import type { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import type {
  SourceConnectionType,
  SourceErrorCodes,
  SourceFileType,
} from "./service/ErrorEventFactory";

export class SourceErrorEventHandler {
  public constructor(
    private readonly metricsService: MetricsService,
    private readonly commonUserMetrics: CommonUserFields
  ) {
    this.commonUserMetrics = commonUserMetrics;
  }

  public fireSourceErrorEvent(
    space: MetricsEventSpace,
    screen_name: MetricsEventScreenName,
    error_code: SourceErrorCodes,
    connection_type: SourceConnectionType,
    file_type: SourceFileType
  ) {
    return this.metricsService.dispatch("sourceErrorEvent", [
      this.commonUserMetrics,
      space,
      screen_name,
      error_code,
      connection_type,
      file_type,
    ]);
  }
}
