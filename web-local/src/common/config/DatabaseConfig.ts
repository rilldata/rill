import { Config, NonFunctionProperties } from "../utils/Config";

const DUCK_MEMORY_DB = ":memory:";

export class DatabaseConfig extends Config<DatabaseConfig> {
  @Config.ConfigField("stage.db")
  public databaseName: string;

  @Config.ConfigField("export")
  public exportFolder: string;

  @Config.ConfigField(false)
  public skipDatabase: boolean;

  @Config.ConfigField()
  public runtimeUrl: string;

  @Config.ConfigField(true)
  public spawnRuntime: boolean;

  @Config.ConfigField(8081)
  public spawnRuntimePort: number;

  constructor(configJson: {
    [K in keyof NonFunctionProperties<DatabaseConfig>]?: NonFunctionProperties<DatabaseConfig>[K];
  }) {
    super(configJson);

    try {
      if (process.env.RILL_EXTERNAL_RUNTIME) {
        this.spawnRuntime = false;
      }
    } catch (err) {
      // no-op
    }
    if (!this.runtimeUrl) {
      this.runtimeUrl = `http://localhost:${this.spawnRuntimePort}`;
    }
  }

  public prependProjectFolder(projectFolder: string) {
    this.exportFolder = `${projectFolder}/${this.exportFolder}`;
    this.databaseName =
      this.databaseName === DUCK_MEMORY_DB
        ? ""
        : `${projectFolder}/${this.databaseName}`;
  }
}
