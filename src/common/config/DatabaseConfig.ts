import {Config} from "$common/utils/Config";

export class DatabaseConfig extends Config<DatabaseConfig> {
    @Config.ConfigField("stage.db")
    public databaseName: string;

    @Config.ConfigField("export")
    public exportFolder: string;

    @Config.ConfigField(false)
    public skipDatabase: boolean;
}
