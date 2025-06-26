import * as fs from "fs";
import * as path from "path";

/**
 * Writes data to a file ensuring that its directory exists.
 *
 * @param filePath - The file path where data should be written.
 * @param data - The data to write.
 * @param options - Optional write options (encoding, flag, etc.).
 */
export function writeFileEnsuringDir(
  filePath: string,
  data: string,
  options?: fs.WriteFileOptions,
): void {
  const dir = path.dirname(filePath);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
  fs.writeFileSync(filePath, data, options);
}
