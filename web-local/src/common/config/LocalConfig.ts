import { Config, NonFunctionProperties } from "../utils/Config";

/**
 * Config that sits locally per install.
 */
export class LocalConfig extends Config<LocalConfig> {
  @Config.ConfigField()
  public installId: string;

  @Config.ConfigField()
  public version: string;

  @Config.ConfigField(false)
  public isDev: boolean;

  @Config.ConfigField(true)
  public sendTelemetryData: boolean;

  public constructor(configJson: {
    [K in keyof NonFunctionProperties<LocalConfig>]?: NonFunctionProperties<LocalConfig>[K];
  }) {
    super(configJson);
    this.setFields(configJson);
  }
}
