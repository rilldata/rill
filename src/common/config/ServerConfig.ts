import {Config} from "$common/utils/Config";

export class ServerConfig extends Config<ServerConfig> {
    @Config.ConfigField("localhost")
    public serverHost: string;
    @Config.ConfigField(8080)
    public serverPort: number;
    public serverUrl: string;

    @Config.ConfigField("localhost")
    public uiHost: string;
    @Config.ConfigField(3000)
    public uiPort: number;
    public uiUrl: string;

    @Config.ConfigField("localhost")
    public socketHost: string;
    @Config.ConfigField(8080)
    public socketPort: number;
    public socketUrl: string;

    @Config.ConfigField(false)
    public serveStaticFile: boolean;

    constructor(configJson) {
        super(configJson);

        this.serverUrl = `http://${this.serverHost}:${this.serverPort}`;
        this.uiUrl = `http://${this.uiHost}:${this.uiPort}`;
        this.socketUrl = `http://${this.socketHost}:${this.socketPort}`;
    }
}
