import type { BehaviourEventMedium } from "@rilldata/web-local/common/metrics-service/BehaviourEventTypes";
import type { MetricsService } from "@rilldata/web-local/common/metrics-service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-local/common/metrics-service/MetricsTypes";

export class NavigationEventHandler {
  public constructor(
    private readonly metricsService: MetricsService,
    private readonly commonUserMetrics: CommonUserFields
  ) {
    this.commonUserMetrics = commonUserMetrics;
  }

  public fireEvent(
    entity_id: string,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    source_screen: MetricsEventScreenName,
    screen_name: MetricsEventScreenName
  ) {
    return this.metricsService.dispatch("navigationEvent", [
      this.commonUserMetrics,
      entity_id,
      medium,
      space,
      source_screen,
      screen_name,
    ]);
  }
}
