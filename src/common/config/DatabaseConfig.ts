import {Config} from "$common/utils/Config";

const DUCK_MEMORY_DB = ":memory:";

export class DatabaseConfig extends Config<DatabaseConfig> {
    @Config.ConfigField("stage.db")
    public databaseName: string;

    @Config.ConfigField("export")
    public exportFolder: string;

    @Config.ConfigField(false)
    public skipDatabase: boolean;

    public prependProjectFolder(projectFolder: string) {
        this.exportFolder =
            `${projectFolder}/${this.exportFolder}`;
        this.databaseName = this.databaseName === DUCK_MEMORY_DB ?
            this.databaseName : `${projectFolder}/${this.databaseName}`;
    }
}
