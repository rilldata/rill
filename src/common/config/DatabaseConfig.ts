import {Config} from "$common/utils/Config";

export class DatabaseConfig extends Config<DatabaseConfig> {
    @Config.ConfigField(":memory:")
    public databaseName: string;

    @Config.ConfigField(".")
    public parquetFolder: string;
}
