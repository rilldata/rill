import {Config} from "$common/utils/Config";
import {ServerConfig} from "$common/config/ServerConfig";
import {DatabaseConfig} from "$common/config/DatabaseConfig";
import { StateConfig } from "$common/config/StateConfig";

export class RootConfig extends Config<RootConfig> {
    @Config.SubConfig(ServerConfig)
    public server: ServerConfig;

    @Config.SubConfig(DatabaseConfig)
    public database: DatabaseConfig;

    @Config.SubConfig(StateConfig)
    public state: StateConfig;
}
