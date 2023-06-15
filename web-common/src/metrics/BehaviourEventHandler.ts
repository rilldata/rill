import type {
  BehaviourEventAction,
  BehaviourEventMedium,
} from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import type { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import MD5 from "crypto-js/md5";
import type {
  SourceConnectionType,
  SourceFileType,
} from "./service/SourceEventTypes";

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

  public firePublishEvent(
    entity_name: string,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    source_screen: MetricsEventScreenName,
    screen_name: MetricsEventScreenName,
    isStart: boolean
  ) {
    const hashedName = MD5(entity_name).toString();
    return this.metricsService.dispatch("publishEvent", [
      this.commonUserMetrics,
      hashedName,
      medium,
      space,
      source_screen,
      screen_name,
      isStart,
    ]);
  }

  public fireSplashEvent(
    action: BehaviourEventAction,
    medium: BehaviourEventMedium,
    space: MetricsEventSpace,
    project_id = ""
  ) {
    return this.metricsService.dispatch("splashEvent", [
      this.commonUserMetrics,
      action,
      medium,
      space,
      project_id,
    ]);
  }

  public fireSourceSuccessEvent(
    medium: BehaviourEventMedium,
    screen_name: MetricsEventScreenName,
    space: MetricsEventSpace,
    connection_type: SourceConnectionType,
    file_type: SourceFileType,
    glob: boolean
  ) {
    return this.metricsService.dispatch("sourceSuccess", [
      this.commonUserMetrics,
      medium,
      screen_name,
      space,
      connection_type,
      file_type,
      glob,
    ]);
  }

  public fireSourceTriggerEvent(
    action: BehaviourEventAction,
    medium: BehaviourEventMedium,
    screen_name: MetricsEventScreenName,
    space: MetricsEventSpace
  ) {
    return this.metricsService.dispatch("sourceTrigger", [
      this.commonUserMetrics,
      action,
      medium,
      screen_name,
      space,
    ]);
  }
}
