/**
 * Script entry point: generates TanStack Query hooks from ConnectRPC
 * service descriptors.
 *
 * Usage: tsx src/runtime-client/v2/codegen/run.ts
 * Output: src/runtime-client/v2/gen/
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { execSync } from "node:child_process";
import { fileURLToPath } from "node:url";

import { QueryService } from "../../../proto/gen/rill/runtime/v1/queries_connect";
import { RuntimeService } from "../../../proto/gen/rill/runtime/v1/api_connect";
import { ConnectorService } from "../../../proto/gen/rill/runtime/v1/connectors_connect";

import {
  generateServiceFile,
  generateIndex,
  loadOrvalTypes,
  toKebabCase,
  type ServiceDef,
} from "./generator";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const outDir = path.resolve(__dirname, "../gen");

fs.mkdirSync(outDir, { recursive: true });

const availableOrvalTypes = loadOrvalTypes(__dirname);

const services: { descriptor: ServiceDef; name: string }[] = [
  { descriptor: QueryService as unknown as ServiceDef, name: "QueryService" },
  {
    descriptor: RuntimeService as unknown as ServiceDef,
    name: "RuntimeService",
  },
  {
    descriptor: ConnectorService as unknown as ServiceDef,
    name: "ConnectorService",
  },
];

let totalQueries = 0;
let totalMutations = 0;

for (const { descriptor, name } of services) {
  const { code, methods } = generateServiceFile(
    descriptor,
    availableOrvalTypes,
  );
  const fileName = `${toKebabCase(name)}.ts`;
  fs.writeFileSync(path.join(outDir, fileName), code);

  const queries = methods.filter((m) => m.classification === "query").length;
  const infinite = methods.filter(
    (m) => m.classification === "query" && m.hasPageToken && m.hasNextPageToken,
  ).length;
  const mutations = methods.filter(
    (m) => m.classification === "mutation",
  ).length;
  const skipped = Object.keys(descriptor.methods).length - methods.length;
  totalQueries += queries;
  totalMutations += mutations;

  console.log(
    `  ${fileName}: ${queries} queries (${infinite} infinite), ${mutations} mutations, ${skipped} skipped`,
  );
}

// Generate barrel index
const indexCode = generateIndex(services.map((s) => s.name));
fs.writeFileSync(path.join(outDir, "index.ts"), indexCode);

console.log(`\nTotal: ${totalQueries} queries, ${totalMutations} mutations`);
console.log(`Output: ${outDir}`);

// Format generated files with prettier
console.log(`\nFormatting with prettier...`);
execSync("npx", ["prettier", "--write", `${outDir}/*.ts`], {
  stdio: "inherit",
});
