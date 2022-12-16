import type { V1RuntimeGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import type {
  ActionServiceBase,
  ExtractActionTypeDefinitions,
  PickActionFunctions,
} from "@rilldata/web-local/lib/metrics/service/ServiceBase";
import { getActionMethods } from "@rilldata/web-local/lib/metrics/service/ServiceBase";
import type { MetricsEventFactory } from "./MetricsEventFactory";
import type { ProductHealthEventFactory } from "./ProductHealthEventFactory";
import type { RillIntakeClient } from "./RillIntakeClient";
import type { CommonFields, MetricsEvent } from "./MetricsTypes";
import type { BehaviourEventFactory } from "./BehaviourEventFactory";
import MD5 from "crypto-js/md5";

/**
 * We have DataModelerStateService as the 1st arg to have a structure for PickActionFunctions
 */
export type MetricsEventFactoryClasses = PickActionFunctions<
  CommonFields,
  ProductHealthEventFactory & BehaviourEventFactory
>;
export type MetricsActionDefinition = ExtractActionTypeDefinitions<
  CommonFields,
  MetricsEventFactoryClasses
>;

export class MetricsService
  implements ActionServiceBase<MetricsActionDefinition>
{
  private actionsMap: {
    [Action in keyof MetricsActionDefinition]?: MetricsEventFactoryClasses;
  } = {};

  private commonFields: Record<string, unknown>;

  public constructor(
    private readonly localConfig: V1RuntimeGetConfig,
    private readonly rillIntakeClient: RillIntakeClient,
    private readonly metricsEventFactories: Array<MetricsEventFactory>
  ) {
    metricsEventFactories.forEach((actions) => {
      getActionMethods(actions).forEach((action) => {
        this.actionsMap[action] = actions;
      });
    });
  }

  public async loadCommonFields() {
    const projectPathParts = this.localConfig.project_path.split("/");
    this.commonFields = {
      app_name: "rill-developer",
      install_id: this.localConfig.install_id,
      build_id: this.localConfig.build_commit,
      version: this.localConfig.version,
      is_dev: this.localConfig.is_dev,
      project_id: MD5(projectPathParts[projectPathParts.length - 1]).toString(),
      analytics_enabled: this.localConfig.analytics_enabled,
    };
  }

  public async dispatch<Action extends keyof MetricsActionDefinition>(
    action: Action,
    args: MetricsActionDefinition[Action]
  ): Promise<void> {
    if (!this.commonFields.analytics_enabled) return;
    if (!this.actionsMap[action]?.[action]) {
      console.log(`${action} not found`);
      return;
    }
    const actionsInstance = this.actionsMap[action];
    const event: MetricsEvent = await actionsInstance[action].call(
      actionsInstance,
      { ...this.commonFields },
      ...args
    );
    await this.rillIntakeClient.fireEvent(event);
  }
}
