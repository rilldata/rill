import type { ActionServiceBase, ExtractActionTypeDefinitions } from "$common/ServiceBase";
import type { PickActionFunctions } from "$common/ServiceBase";
import type { RootConfig } from "$common/config/RootConfig";
import type { CommonMetricsFields, MetricsEvent } from "$common/metrics/MetricsTypes";
import type { MetricsEventFactory } from "$common/metrics/MetricsEventFactory";
import type { ProductHealthEventFactory } from "$common/metrics/ProductHealthEventFactory";
import { getActionMethods } from "$common/ServiceBase";
import type { RillIntakeClient } from "$common/metrics/RillIntakeClient";

export type MetricsEventFactoryClasses = PickActionFunctions<CommonMetricsFields, (
    ProductHealthEventFactory
)>;
export type MetricsActionDefinition = ExtractActionTypeDefinitions<CommonMetricsFields, MetricsEventFactoryClasses>;

export class MetricsService implements ActionServiceBase<MetricsActionDefinition> {
    private actionsMap: {
        [Action in keyof MetricsActionDefinition]?: MetricsEventFactoryClasses
    } = {};

    public constructor(private readonly config: RootConfig,
                       private commonMetricsInput: CommonMetricsFields,
                       private readonly rillIntakeClient: RillIntakeClient,
                       private readonly metricsEventFactories: Array<MetricsEventFactory>) {
        metricsEventFactories.forEach((actions) => {
            getActionMethods(actions).forEach(action => {
                this.actionsMap[action] = actions;
            });
        });
    }

    public setCommonMetricsInput(commonMetricsInput: CommonMetricsFields) {
        this.commonMetricsInput = commonMetricsInput;
    }

    public async dispatch<Action extends keyof MetricsActionDefinition>(
        action: Action, args: MetricsActionDefinition[Action],
    ): Promise<any> {
        if (!this.actionsMap[action]?.[action]) {
            console.log(`${action} not found`);
            return;
        }
        const actionsInstance = this.actionsMap[action];
        const event: MetricsEvent = await actionsInstance[action].call(actionsInstance,
            this.commonMetricsInput, ...args);
        await this.rillIntakeClient.fireEvent(event);
    }
}
