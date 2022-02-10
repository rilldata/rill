import { Config } from "$common/utils/Config";

export class StateConfig extends Config<StateConfig> {
    @Config.ConfigField("saved-state.json")
    public savedStateFile: string;
}
