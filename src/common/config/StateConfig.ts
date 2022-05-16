import { Config } from "$common/utils/Config";

export class StateConfig extends Config<StateConfig> {
  @Config.ConfigField(true)
  public autoSync: boolean;

  @Config.ConfigField(500)
  public syncInterval: number;

  @Config.ConfigField("state")
  public stateFolder: string;

  @Config.ConfigField("models")
  public modelFolder: string;

  public prependProjectFolder(projectFolder: string) {
    this.stateFolder = `${projectFolder}/${this.stateFolder}`;
    this.modelFolder = `${projectFolder}/${this.modelFolder}`;
  }
}
