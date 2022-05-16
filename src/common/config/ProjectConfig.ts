import { Config } from "$common/utils/Config";

export class ProjectConfig extends Config<ProjectConfig> {
  @Config.ConfigField()
  public duckDbPath: string;
}
