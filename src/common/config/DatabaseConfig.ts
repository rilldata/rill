import {Config} from "$common/utils/Config";

export class DatabaseConfig extends Config<DatabaseConfig> {
    @Config.ConfigField("project.db")
    public databaseName: string;

    @Config.ConfigField(".")
    public parquetFolder: string;

    @Config.ConfigField("export")
    public exportFolder: string;
}
