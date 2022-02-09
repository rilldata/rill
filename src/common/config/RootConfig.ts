import {Config} from "$common/utils/Config";
import {ServerConfig} from "$common/config/ServerConfig";
import {DatabaseConfig} from "$common/config/DatabaseConfig";

export class RootConfig extends Config<RootConfig> {
    @Config.SubConfig(ServerConfig)
    public server: ServerConfig;

    @Config.SubConfig(DatabaseConfig)
    public database: DatabaseConfig;
}
