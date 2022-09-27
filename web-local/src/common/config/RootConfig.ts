import { Config } from "../utils/Config";
import type { NonFunctionProperties } from "../utils/Config";
import { ServerConfig } from "./ServerConfig";
import { DatabaseConfig } from "./DatabaseConfig";
import { StateConfig } from "./StateConfig";
import { MetricsConfig } from "./MetricsConfig";
import { LocalConfig } from "./LocalConfig";
import { ProjectConfig } from "./ProjectConfig";

export class RootConfig extends Config<RootConfig> {
  @Config.SubConfig(ServerConfig)
  public server: ServerConfig;

  @Config.SubConfig(DatabaseConfig)
  public database: DatabaseConfig;

  @Config.SubConfig(StateConfig)
  public state: StateConfig;

  @Config.SubConfig(MetricsConfig)
  public metrics: MetricsConfig;

  @Config.SubConfig(LocalConfig)
  public local: LocalConfig;

  @Config.SubConfig(ProjectConfig)
  public project: ProjectConfig;

  @Config.ConfigField(".")
  public projectFolder: string;

  @Config.ConfigField(true)
  public profileWithUpdate: boolean;

  constructor(configJson: {
    [K in keyof NonFunctionProperties<RootConfig>]?: NonFunctionProperties<RootConfig>[K];
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
