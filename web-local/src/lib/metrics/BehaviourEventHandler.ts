import type {
  BehaviourEventMedium,
  BehaviourEventAction,
} from "@rilldata/web-local/lib/metrics/service/BehaviourEventTypes";
import type { MetricsService } from "@rilldata/web-local/lib/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-local/lib/metrics/service/MetricsTypes";
import MD5 from "crypto-js/md5";

// TODO: simplify telemetry code to fewer classes and layers
export class BehaviourEventHandler {
  public constructor(
    private readonly metricsService: MetricsService,
    private readonly commonUserMetrics: CommonUserFields
  ) {
    this.commonUserMetrics = commonUserMetrics;
  }

  public fireNavigationEvent(
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

  public fireDeployEvent(project_name: string, action: BehaviourEventAction) {
    const project_id = MD5(project_name).toString();
    return this.metricsService.dispatch("deployEvent", [
      this.commonUserMetrics,
      project_id,
      action,
    ]);
  }
}
