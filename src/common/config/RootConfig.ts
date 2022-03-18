import { Config } from "$common/utils/Config";
import type { NonFunctionProperties } from "$common/utils/Config";
import {ServerConfig} from "$common/config/ServerConfig";
import {DatabaseConfig} from "$common/config/DatabaseConfig";
import { StateConfig } from "$common/config/StateConfig";
import { MetricsConfig } from "$common/config/MetricsConfig";

export class RootConfig extends Config<RootConfig> {
    @Config.SubConfig(ServerConfig)
    public server: ServerConfig;

    @Config.SubConfig(DatabaseConfig)
    public database: DatabaseConfig;

    @Config.SubConfig(StateConfig)
    public state: StateConfig;

    @Config.SubConfig(MetricsConfig)
    public metrics: MetricsConfig;

    @Config.ConfigField(".")
    public projectFolder: string;

    @Config.ConfigField(true)
    public profileWithUpdate: boolean;

    constructor(configJson: {
        [K in keyof NonFunctionProperties<RootConfig>]?: NonFunctionProperties<RootConfig>[K]
    }) {
        super(configJson);

        this.prependProjectFolder();
    }

    private prependProjectFolder() {
        if (this.projectFolder === ".") return;

        this.database.prependProjectFolder(this.projectFolder);
        this.state.prependProjectFolder(this.projectFolder);
    }
}
