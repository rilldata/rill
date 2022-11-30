import { Config, NonFunctionProperties } from "../utils/Config";

export class StateConfig extends Config<StateConfig> {
  @Config.ConfigField(true)
  public autoSync: boolean;

  @Config.ConfigField(500)
  public syncInterval: number;

  @Config.ConfigField("state")
  public stateFolder: string;

  @Config.ConfigField("models")
  public modelFolder: string;

  public constructor(configJson: {
    [K in keyof NonFunctionProperties<StateConfig>]?: NonFunctionProperties<StateConfig>[K];
  }) {
    super(configJson);
    this.setFields(configJson);
  }

  public prependProjectFolder(projectFolder: string) {
    this.stateFolder = `${projectFolder}/${this.stateFolder}`;
    this.modelFolder = `${projectFolder}/${this.modelFolder}`;
  }
}
