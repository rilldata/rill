import { ApplicationConfigFolder } from "../config/ConfigFolders";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
  chmodSync,
} from "fs";
import path from "path";

const RuntimeTempFolder = `${ApplicationConfigFolder}/temp`;

export function getBinaryRuntimePath(version: string) {
  let runtimeBinaryPath = path.join(__dirname, "/../../runtime/runtime");

  if (!existsSync(runtimeBinaryPath)) {
    runtimeBinaryPath = path.join(__dirname, "/../../../dist/runtime/runtime");
  }

  // fix for vercel treating the runtime executable as an asset
  if (runtimeBinaryPath.startsWith("/snapshot")) {
    const newBinaryPath = `${RuntimeTempFolder}/${version}/runtime`;
    // if the runtime binary for version doesnt exist,
    // copy over the content to an external file.
    // this is because the binary that is treated as an asset is not actually persisted to a file
    if (!existsSync(newBinaryPath)) {
      mkdirSync(`${RuntimeTempFolder}/${version}`, {
        recursive: true,
      });
      writeFileSync(newBinaryPath, readFileSync(runtimeBinaryPath));
      chmodSync(newBinaryPath, 755);
    }
    return newBinaryPath;
  }

  return runtimeBinaryPath.replace(/ /g, "\\ ");
}
