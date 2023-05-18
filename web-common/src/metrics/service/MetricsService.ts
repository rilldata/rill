import type {
  ActionServiceBase,
  ExtractActionTypeDefinitions,
  PickActionFunctions,
} from "@rilldata/web-common/metrics/service/ServiceBase";
import { getActionMethods } from "@rilldata/web-common/metrics/service/ServiceBase";
import type { V1RuntimeGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import MD5 from "crypto-js/md5";
import { v4 as uuidv4 } from "uuid";
import type { BehaviourEventFactory } from "./BehaviourEventFactory";
import type { MetricsEventFactory } from "./MetricsEventFactory";
import type { CommonFields, MetricsEvent } from "./MetricsTypes";
import type { ProductHealthEventFactory } from "./ProductHealthEventFactory";
import type { RillIntakeClient } from "./RillIntakeClient";

export const ClientIDStorageKey = "client_id";

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
      client_id: this.getOrSetClientID(),
      build_id: this.localConfig.build_commit,
      version: this.localConfig.version,
      is_dev: this.localConfig.is_dev,
      project_id: MD5(projectPathParts[projectPathParts.length - 1]).toString(),
      user_id: this.localConfig.user_id,
      analytics_enabled: this.localConfig.analytics_enabled,
      mode: this.localConfig.readonly ? "read-only" : "edit",
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

  private getOrSetClientID(): string {
    let clientId = localStorage.getItem(ClientIDStorageKey);
    if (clientId) return clientId;

    clientId = uuidv4();
    localStorage.setItem(ClientIDStorageKey, clientId);
    return clientId;
  }
}
