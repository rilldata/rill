import { Config } from "$common/utils/Config";

const ApplicationConfigFolder = "~/.rill";
const LocalConfigFile = `${ApplicationConfigFolder}/local.json`;

/**
 * Config that sits locally per install.
 */
export class LocalConfig extends Config<LocalConfig> {
    @Config.ConfigField()
    public installId: string;

    @Config.ConfigField(true)
    public sendTelemetryData: boolean;
}
