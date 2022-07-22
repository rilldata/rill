import path from "path";
import { fileURLToPath } from "url";

// @ts-ignore
const fileName = fileURLToPath(import.meta.url);
const dirName = path.dirname(fileName);
export { dirName as __dirname };
