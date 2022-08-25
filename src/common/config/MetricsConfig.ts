import { Config } from "$common/utils/Config";

export class MetricsConfig extends Config<MetricsConfig> {
  @Config.ConfigField("rill-developer")
  public appName: string;

  @Config.ConfigField(60)
  public activeEventInterval: number;

  @Config.ConfigField("https://intake.rilldata.io/events/data-modeler-metrics")
  public rillIntakeUrl: string;

  @Config.ConfigField("data-modeler")
  public rillIntakeUser: string;

  @Config.ConfigField(
    "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug=="
  )
  public rillIntakePassword: string;
}
