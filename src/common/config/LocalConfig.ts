import { Config } from "$common/utils/Config";

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
}
