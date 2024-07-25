import { readFileSync } from "fs";

export function readPublicEmailDomains() {
  const contents = readFileSync(
    __dirname +
      "/../../../../../admin/pkg/publicemail/public_email_providers_list",
  ).toString();
  return contents
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => !l.startsWith("#"));
}
