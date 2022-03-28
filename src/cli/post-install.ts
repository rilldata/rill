import "../moduleAlias";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "fs";
import { LocalConfig} from "$common/config/LocalConfig";
import { guidGenerator } from "$lib/util/guid";
import { ApplicationConfigFolder, LocalConfigFile } from "$common/config/ConfigFolders";

/**
 * Initializes the rill local config.
 * 1. Creates a folder under ApplicationConfigFolder
 * 2. Creates a config file LocalConfigFile
 * 3. Generates installId and saves into the LocalConfigFile
 *
 * This is run as postinstall npm script.
 */

(async () => {
    if (!existsSync(ApplicationConfigFolder)) {
        mkdirSync(ApplicationConfigFolder, {recursive: true});
        console.log("creating folder");
    }

    let configJson;
    if (existsSync(LocalConfigFile)) {
        configJson = JSON.parse(readFileSync(LocalConfigFile).toString());
    } else {
        configJson = {};
    }
    const configObject = new LocalConfig(configJson);
    configObject.installId = guidGenerator();

    // We should instead move this to UI
    // configObject.sendTelemetryData = await cliConfirmation(
    //     "We collect usage info to improve Rill Developer. " +
    //     "Send anonymous tracking data? (y/N)");

    writeFileSync(LocalConfigFile, JSON.stringify(configObject));
})();
