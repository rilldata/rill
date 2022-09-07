import { existsSync, mkdirSync, readFileSync, writeFileSync } from "fs";
import {
  ApplicationConfigFolder,
  LocalConfigFile,
} from "$common/config/ConfigFolders";
import { guidGenerator } from "$lib/util/guid";
import { LocalConfig } from "$common/config/LocalConfig";

/**
 * Initializes the rill local config.
 * 1. Creates a folder under ApplicationConfigFolder
 * 2. Creates a config file LocalConfigFile
 * 3. Generates installId and saves into the LocalConfigFile
 */
export function initLocalConfig(localConfig?: LocalConfig) {
  if (!existsSync(ApplicationConfigFolder)) {
    mkdirSync(ApplicationConfigFolder, { recursive: true });
    console.log("creating folder");
  }

  let configJson;

  if (existsSync(LocalConfigFile)) {
    try {
      configJson = JSON.parse(readFileSync(LocalConfigFile).toString());
    } catch (err) {
      console.error("Error reading local config.");
    }
  }
  if (!configJson) {
    // generate install id only for the 1st time
    configJson = {
      installId: guidGenerator(),
    };
    writeFileSync(LocalConfigFile, JSON.stringify(configJson));
  }

  if (localConfig?.isDev) {
    configJson.isDev = localConfig.isDev;
  }

  return new LocalConfig(configJson);
}
