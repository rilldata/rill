import { Config, NonFunctionProperties } from "../utils/Config";

export class ProjectConfig extends Config<ProjectConfig> {
  @Config.ConfigField()
  public duckDbPath: string;

  public constructor(configJson: {
    [K in keyof NonFunctionProperties<ProjectConfig>]?: NonFunctionProperties<ProjectConfig>[K];
  }) {
    super(configJson);
    this.setFields(configJson);
  }
}
