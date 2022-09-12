import type {
  ActionServiceBase,
  ExtractActionTypeDefinitions,
  PickActionFunctions,
} from "$common/ServiceBase";
import { getActionMethods } from "$common/ServiceBase";
import type { RootConfig } from "$common/config/RootConfig";
import type { MetricsEventFactory } from "$common/metrics-service/MetricsEventFactory";
import type { ProductHealthEventFactory } from "$common/metrics-service/ProductHealthEventFactory";
import type { RillIntakeClient } from "$common/metrics-service/RillIntakeClient";
import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type {
  CommonFields,
  MetricsEvent,
} from "$common/metrics-service/MetricsTypes";
import type { BehaviourEventFactory } from "$common/metrics-service/BehaviourEventFactory";

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

  public constructor(
    private readonly config: RootConfig,
    private readonly dataModelerStateService: DataModelerStateService,
    private readonly rillIntakeClient: RillIntakeClient,
    private readonly metricsEventFactories: Array<MetricsEventFactory>
  ) {
    metricsEventFactories.forEach((actions) => {
      getActionMethods(actions).forEach((action) => {
        this.actionsMap[action] = actions;
      });
    });
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
      this.getCommonFields(),
      ...args
    );
    await this.rillIntakeClient.fireEvent(event);
  }

  private getCommonFields(): CommonFields {
    const applicationState = this.dataModelerStateService.getApplicationState();
    return {
      app_name: this.config.metrics.appName,
      install_id: this.config.local.installId,
      build_id: this.config.local.version ?? "",
      version: this.config.local.version ?? "",
      is_dev: this.config.local.isDev,
      project_id: applicationState.projectId,
    };
  }
}
