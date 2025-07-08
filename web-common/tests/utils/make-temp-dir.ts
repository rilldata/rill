import { mkdtempSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";

export function makeTempDir(dirName: string) {
  return mkdtempSync(join(tmpdir(), dirName));
}
