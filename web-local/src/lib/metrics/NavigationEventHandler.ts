import type { BehaviourEventMedium } from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
import type { MetricsService } from "@rilldata/web-local/lib/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
import MD5 from "crypto-js/md5";

export class NavigationEventHandler {
  public constructor(
    private readonly metricsService: MetricsService,
    private readonly commonUserMetrics: CommonUserFields
  ) {
    this.commonUserMetrics = commonUserMetrics;
  }

  public fireEvent(
    entity_name: string,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    source_screen: MetricsEventScreenName,
    screen_name: MetricsEventScreenName
  ) {
    const hashedName = MD5(entity_name).toString();
    return this.metricsService.dispatch("navigationEvent", [
      this.commonUserMetrics,
      hashedName,
      medium,
      space,
      source_screen,
      screen_name,
    ]);
  }
}
