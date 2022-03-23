import "../moduleAlias";
import { existsSync, mkdirSync, readFileSync, writeFileSync } from "fs";
import { LocalConfig} from "$common/config/LocalConfig";
import { guidGenerator } from "$lib/util/guid";
import { ApplicationConfigFolder, LocalConfigFile } from "$common/config/ConfigFolders";

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

writeFileSync(LocalConfigFile, JSON.stringify(configObject));
