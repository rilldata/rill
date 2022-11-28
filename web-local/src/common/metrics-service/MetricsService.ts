import { fetchWrapperDirect } from "@rilldata/web-local/lib/util/fetchWrapper";
import type {
  ActionServiceBase,
  ExtractActionTypeDefinitions,
  PickActionFunctions,
} from "../ServiceBase";
import { getActionMethods } from "../ServiceBase";
import type { RootConfig } from "../config/RootConfig";
import type { MetricsEventFactory } from "./MetricsEventFactory";
import type { ProductHealthEventFactory } from "./ProductHealthEventFactory";
import type { RillIntakeClient } from "./RillIntakeClient";
import type { CommonFields, MetricsEvent } from "./MetricsTypes";
import type { BehaviourEventFactory } from "./BehaviourEventFactory";

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
    private readonly config: RootConfig,
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
    const localConfig = await fetchWrapperDirect(
      `${this.config.server.serverUrl}/local/config`,
      "GET"
    );
    this.commonFields = {
      app_name: this.config.metrics.appName,
      install_id: localConfig.install_id,
      // @ts-ignore
      build_id: RILL_COMMIT,
      // @ts-ignore
      version: RILL_VERSION,
      is_dev: localConfig.is_dev,
      project_id: localConfig.project_id,
    };
  }

  public async dispatch<Action extends keyof MetricsActionDefinition>(
    action: Action,
    args: MetricsActionDefinition[Action]
  ): Promise<void> {
    if (!this.config.local.sendTelemetryData) return;
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
