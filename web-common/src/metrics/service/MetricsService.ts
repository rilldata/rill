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
import type { ErrorEventFactory } from "./ErrorEventFactory";
import type { CommonFields, MetricsEvent } from "./MetricsTypes";
import type { ProductHealthEventFactory } from "./ProductHealthEventFactory";
import type { TelemetryClient } from "./RillIntakeClient";

export const ClientIDStorageKey = "client_id";

/**
 * We have DataModelerStateService as the 1st arg to have a structure for PickActionFunctions
 */
export type MetricsEventFactoryClasses = PickActionFunctions<
  CommonFields,
  ProductHealthEventFactory & BehaviourEventFactory & ErrorEventFactory
>;
export type MetricsActionDefinition = ExtractActionTypeDefinitions<
  CommonFields,
  MetricsEventFactoryClasses
>;

export interface CloudMetricsFields {
  isDev: boolean;
  projectId: string;
  organizationId: string;
  userId: string;
}

export class MetricsService
  implements ActionServiceBase<MetricsActionDefinition>
{
  private actionsMap: {
    [Action in keyof MetricsActionDefinition]?: MetricsEventFactoryClasses;
  } = {};

  private commonFields: Record<string, unknown>;

  public constructor(
    private readonly telemetryClient: TelemetryClient,
    metricsEventFactories: Array<MetricsEventFactory>,
  ) {
    metricsEventFactories.forEach((actions) => {
      getActionMethods(actions).forEach((action) => {
        this.actionsMap[action] = actions;
      });
    });
  }

  public loadLocalFields(localConfig: V1RuntimeGetConfig) {
    const projectPathParts = localConfig.project_path.split("/");
    this.commonFields = {
      service_name: "web-local",
      app_name: "rill-developer",
      install_id: localConfig.install_id,
      client_id: this.getOrSetClientID(),
      build_id: localConfig.build_commit,
      version: localConfig.version,
      is_dev: localConfig.is_dev,
      project_id: MD5(projectPathParts[projectPathParts.length - 1]).toString(),
      user_id: localConfig.user_id,
      analytics_enabled: localConfig.analytics_enabled,
      mode: localConfig.readonly ? "read-only" : "edit",
    };
  }

  public loadCloudFields(fields: CloudMetricsFields) {
    this.commonFields = {
      // TODO: build_id and version
      service_name: "web-admin",
      app_name: "rill-cloud",
      client_id: this.getOrSetClientID(),
      is_dev: fields.isDev,
      project_id: fields.projectId,
      organization_id: fields.organizationId,
      user_id: fields.userId,
      analytics_enabled: true,
    };
  }

  public async dispatch<Action extends keyof MetricsActionDefinition>(
    action: Action,
    args: MetricsActionDefinition[Action],
  ): Promise<void> {
    if (!this.commonFields?.analytics_enabled) return;
    const actionsInstance = this.actionsMap[action];
    if (!actionsInstance?.[action]) {
      console.log(`${action} not found`);
      return;
    }
    const event: MetricsEvent = await actionsInstance[action].call(
      actionsInstance,
      { ...this.commonFields },
      ...args,
    );
    await this.telemetryClient.fireEvent(event);
  }

  private getOrSetClientID(): string {
    let clientId = localStorage.getItem(ClientIDStorageKey);
    if (clientId) return clientId;

    clientId = uuidv4() as string;
    localStorage.setItem(ClientIDStorageKey, clientId);
    return clientId;
  }
}
