import { Config } from "$common/utils/Config";

export class MetricsConfig extends Config<MetricsConfig> {
    @Config.ConfigField("data-modeler")
    public appName: string;

    @Config.ConfigField("http://")
    public rillIntakeUrl: string;
}
