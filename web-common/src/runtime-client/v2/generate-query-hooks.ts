/**
 * Code generator: reads ConnectRPC *_connect.ts service descriptors and
 * produces TanStack Query hooks for Svelte.
 *
 * Usage: tsx src/runtime-client/v2/generate-query-hooks.ts
 *
 * Output: src/runtime-client/v2/gen/{query-service,runtime-service,connector-service}.ts
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
import { MethodKind } from "@bufbuild/protobuf";
import {
  classifyMethod,
  type MethodClassification,
} from "./generate-query-hooks-config";

// --- Service descriptors ---

import { QueryService } from "../../proto/gen/rill/runtime/v1/queries_connect";
import { RuntimeService } from "../../proto/gen/rill/runtime/v1/api_connect";
import { ConnectorService } from "../../proto/gen/rill/runtime/v1/connectors_connect";

interface ServiceDef {
  typeName: string;
  methods: Record<
    string,
    {
      name: string;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      I: { typeName: string; new (): any };
      O: { typeName: string };
      kind: MethodKind;
    }
  >;
}

interface MethodInfo {
  /** camelCase method name as it appears in the service descriptor (e.g. "metricsViewAggregation") */
  methodKey: string;
  /** PascalCase method name from the proto (e.g. "MetricsViewAggregation") */
  methodName: string;
  /** Request message type name (e.g. "MetricsViewAggregationRequest") */
  inputType: string;
  /** Response message type name (e.g. "MetricsViewAggregationResponse") */
  outputType: string;
  classification: MethodClassification;
  /** Whether the request type has an instanceId field */
  hasInstanceId: boolean;
}

// --- Helpers ---

function pascalCase(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function extractShortName(typeName: string): string {
  // "rill.runtime.v1.MetricsViewAggregationRequest" -> "MetricsViewAggregationRequest"
  const parts = typeName.split(".");
  return parts[parts.length - 1];
}

function getServiceShortName(typeName: string): string {
  // "rill.runtime.v1.QueryService" -> "QueryService"
  return extractShortName(typeName);
}

function getServiceClientProp(serviceName: string): string {
  // "QueryService" -> "queryService", "RuntimeService" -> "runtimeService"
  return serviceName.charAt(0).toLowerCase() + serviceName.slice(1);
}

/** kebab-case for output filename */
function toKebabCase(s: string): string {
  return s
    .replace(/([a-z])([A-Z])/g, "$1-$2")
    .replace(/([A-Z])([A-Z][a-z])/g, "$1-$2")
    .toLowerCase();
}

/** Derive the _pb.ts module path that exports the request/response types */
function getProtoImportPath(serviceName: string): string {
  const fileMap: Record<string, string> = {
    QueryService: "queries_pb",
    RuntimeService: "api_pb",
    ConnectorService: "connectors_pb",
  };
  return `../../../proto/gen/rill/runtime/v1/${fileMap[serviceName]}`;
}

// --- JSON bridge: Orval type helpers ---

/** Read Orval schemas to discover available V1 types at generation time */
const orvalSchemaPath = path.resolve(__dirname, "../gen/index.schemas.ts");
const orvalSchemaContent = fs.readFileSync(orvalSchemaPath, "utf-8");
const availableOrvalTypes = new Set<string>();
for (const match of orvalSchemaContent.matchAll(
  /^export (?:type|interface) (\w+)/gm,
)) {
  availableOrvalTypes.add(match[1]);
}

const ORVAL_IMPORT_PATH = "../../gen/index.schemas";

function orvalTypeName(protoTypeName: string): string {
  return `V1${protoTypeName}`;
}

function hasOrvalType(protoTypeName: string): boolean {
  return availableOrvalTypes.has(orvalTypeName(protoTypeName));
}

/** Get the public-facing type for a request or response */
function publicType(protoTypeName: string): string {
  return hasOrvalType(protoTypeName)
    ? orvalTypeName(protoTypeName)
    : `PartialMessage<${protoTypeName}>`;
}

// --- Method extraction ---

function extractMethods(service: ServiceDef): MethodInfo[] {
  const serviceName = getServiceShortName(service.typeName);
  const methods: MethodInfo[] = [];

  for (const [key, method] of Object.entries(service.methods)) {
    if (method.kind !== MethodKind.Unary) {
      // Skip streaming methods automatically
      continue;
    }

    const classification = classifyMethod(serviceName, key);
    if (classification === "skip") continue;

    const hasInstanceId = "instanceId" in new method.I();

    methods.push({
      methodKey: key,
      methodName: method.name,
      inputType: extractShortName(method.I.typeName),
      outputType: extractShortName(method.O.typeName),
      classification,
      hasInstanceId,
    });
  }

  return methods;
}

// --- Code generation ---

function generateServiceFile(service: ServiceDef): string {
  const serviceName = getServiceShortName(service.typeName);
  const serviceClientProp = getServiceClientProp(serviceName);
  const protoImportPath = getProtoImportPath(serviceName);
  const methods = extractMethods(service);

  const queryMethods = methods.filter((m) => m.classification === "query");
  const mutationMethods = methods.filter(
    (m) => m.classification === "mutation",
  );

  // Collect proto types (value imports; needed for fromJson/toJson calls)
  const protoTypes = new Set<string>();
  // Collect Orval types (type-only imports; used in public API signatures)
  const orvalTypes = new Set<string>();
  // Track whether PartialMessage is needed (for request types without V1 counterparts)
  let needsPartialMessage = false;

  for (const m of methods) {
    protoTypes.add(m.inputType);

    if (hasOrvalType(m.inputType)) {
      orvalTypes.add(orvalTypeName(m.inputType));
    } else {
      needsPartialMessage = true;
      // Still need proto type for PartialMessage<> reference
    }

    if (hasOrvalType(m.outputType)) {
      orvalTypes.add(orvalTypeName(m.outputType));
    }
  }

  const lines: string[] = [];

  // Header
  lines.push(`// Generated by generate-query-hooks.ts — DO NOT EDIT`);
  lines.push(``);
  lines.push(`import type { RuntimeClient } from "../runtime-client";`);

  // @bufbuild/protobuf imports (JsonValue always needed; PartialMessage conditional)
  const bufImports: string[] = ["JsonValue"];
  if (needsPartialMessage) bufImports.unshift("PartialMessage");
  lines.push(
    `import type { ${bufImports.join(", ")} } from "@bufbuild/protobuf";`,
  );

  lines.push(
    `import {`,
    `  createQuery,`,
    `  createMutation,`,
    `  type CreateQueryOptions,`,
    `  type CreateQueryResult,`,
    `  type QueryClient,`,
    `  type QueryFunction,`,
    `  type QueryKey,`,
    `  type CreateMutationOptions,`,
    `  type CreateMutationResult,`,
    `} from "@tanstack/svelte-query";`,
  );

  // Orval type imports (type-only; for public API signatures)
  if (orvalTypes.size > 0) {
    const sortedOrval = [...orvalTypes].sort();
    lines.push(
      `import type {`,
      ...sortedOrval.map(
        (t, i) => `  ${t}${i < sortedOrval.length - 1 ? "," : ""}`,
      ),
      `} from "${ORVAL_IMPORT_PATH}";`,
    );
  }

  // Proto type imports (value imports; needed for fromJson calls)
  const sortedProto = [...protoTypes].sort();
  lines.push(
    `import {`,
    ...sortedProto.map(
      (t, i) => `  ${t}${i < sortedProto.length - 1 ? "," : ""}`,
    ),
    `} from "${protoImportPath}";`,
  );

  // Utility: strip undefined values before passing to proto fromJson
  // (proto fromJson rejects undefined; Orval's HTTP client silently omitted them)
  lines.push(
    ``,
    `/** Deep-strip undefined values — proto fromJson rejects them */`,
    `// eslint-disable-next-line @typescript-eslint/no-explicit-any`,
    `function stripUndefined(obj: Record<string, any>): Record<string, unknown> {`,
    `  const result: Record<string, unknown> = {};`,
    `  for (const [key, value] of Object.entries(obj)) {`,
    `    if (value === undefined) continue;`,
    `    if (Array.isArray(value)) {`,
    `      result[key] = value.map((item) =>`,
    `        item && typeof item === "object" && !Array.isArray(item)`,
    `          ? stripUndefined(item)`,
    `          : item,`,
    `      );`,
    `    } else if (value && typeof value === "object" && !(value instanceof Date)) {`,
    `      result[key] = stripUndefined(value);`,
    `    } else {`,
    `      result[key] = value;`,
    `    }`,
    `  }`,
    `  return result;`,
    `}`,
  );

  lines.push(``);

  // --- Query methods ---
  for (const m of queryMethods) {
    const fullName = `${serviceClientProp}${pascalCase(m.methodKey)}`;
    const keyFnName = `get${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}QueryKey`;
    const optsFnName = `get${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}QueryOptions`;
    const hookName = `create${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}`;

    // Public-facing request type (V1 Orval when available, PartialMessage fallback)
    const inputPublic = publicType(m.inputType);
    const requestType = m.hasInstanceId
      ? `Omit<${inputPublic}, "instanceId">`
      : inputPublic;
    const requestSpread = m.hasInstanceId
      ? `{ instanceId: client.instanceId, ...request }`
      : `request`;

    // Public-facing response type (always V1 Orval)
    const responseType = publicType(m.outputType);

    // Tier 1: Raw function (JSON bridge: fromJson on input, toJson on output)
    lines.push(
      `/**`,
      ` * Raw RPC call: ${serviceName}.${m.methodName}`,
      ` */`,
      `export async function ${fullName}(`,
      `  client: RuntimeClient,`,
      `  request: ${requestType},`,
      `  options?: { signal?: AbortSignal },`,
      `): Promise<${responseType}> {`,
      `  const r = await client.${serviceClientProp}.${m.methodKey}(`,
      `    ${m.inputType}.fromJson(stripUndefined(${requestSpread}) as unknown as JsonValue),`,
      `    { signal: options?.signal },`,
      `  );`,
      `  return r.toJson({ emitDefaultValues: true }) as unknown as ${responseType};`,
      `}`,
      ``,
    );

    // Tier 2: Query key
    lines.push(
      `export function ${keyFnName}(`,
      `  instanceId: string,`,
      `  request?: ${requestType},`,
      `): QueryKey {`,
      `  return ["${serviceName}", "${m.methodKey}", instanceId, request ?? {}] as const;`,
      `}`,
      ``,
    );

    // Tier 3: Query options (generic TData for select support)
    lines.push(
      `export function ${optsFnName}<TData = ${responseType}>(`,
      `  client: RuntimeClient,`,
      `  request: ${requestType},`,
      `  options?: {`,
      `    query?: Partial<CreateQueryOptions<${responseType}, Error, TData>>;`,
      `  },`,
      `): CreateQueryOptions<${responseType}, Error, TData> & { queryKey: QueryKey } {`,
      `  const queryKey = ${keyFnName}(client.instanceId, request);`,
      `  const queryFn: QueryFunction<${responseType}> = ({ signal }) =>`,
      `    ${fullName}(client, request, { signal });`,
      `  return {`,
      `    queryKey,`,
      `    queryFn,`,
      `    enabled: !!client.instanceId,`,
      `    ...options?.query,`,
      `  } as CreateQueryOptions<${responseType}, Error, TData> & { queryKey: QueryKey };`,
      `}`,
      ``,
    );

    // Tier 4: Convenience hook (generic TData for select support)
    lines.push(
      `export function ${hookName}<TData = ${responseType}>(`,
      `  client: RuntimeClient,`,
      `  request: ${requestType},`,
      `  options?: {`,
      `    query?: Partial<CreateQueryOptions<${responseType}, Error, TData>>;`,
      `  },`,
      `  queryClient?: QueryClient,`,
      `): CreateQueryResult<TData, Error> {`,
      `  const queryOptions = ${optsFnName}(client, request, options);`,
      `  return createQuery(queryOptions, queryClient);`,
      `}`,
      ``,
    );
  }

  // --- Mutation methods ---
  for (const m of mutationMethods) {
    const fullName = `${serviceClientProp}${pascalCase(m.methodKey)}`;
    const mutOptsFnName = `get${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}MutationOptions`;
    const mutHookName = `create${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}Mutation`;

    // Public-facing request type (V1 Orval when available, PartialMessage fallback)
    const inputPublic = publicType(m.inputType);
    const requestType = m.hasInstanceId
      ? `Omit<${inputPublic}, "instanceId">`
      : inputPublic;
    const requestSpread = m.hasInstanceId
      ? `{ instanceId: client.instanceId, ...request }`
      : `request`;

    // Public-facing response type (always V1 Orval)
    const responseType = publicType(m.outputType);

    // Raw function (JSON bridge: fromJson on input, toJson on output)
    lines.push(
      `/**`,
      ` * Raw RPC call (mutation): ${serviceName}.${m.methodName}`,
      ` */`,
      `export async function ${fullName}(`,
      `  client: RuntimeClient,`,
      `  request: ${requestType},`,
      `): Promise<${responseType}> {`,
      `  const r = await client.${serviceClientProp}.${m.methodKey}(`,
      `    ${m.inputType}.fromJson(stripUndefined(${requestSpread}) as unknown as JsonValue),`,
      `  );`,
      `  return r.toJson({ emitDefaultValues: true }) as unknown as ${responseType};`,
      `}`,
      ``,
    );

    // Mutation options
    lines.push(
      `export function ${mutOptsFnName}(`,
      `  client: RuntimeClient,`,
      `  options?: Partial<CreateMutationOptions<${responseType}, unknown, ${requestType}>>,`,
      `): CreateMutationOptions<${responseType}, unknown, ${requestType}> {`,
      `  return {`,
      `    mutationFn: (request) => ${fullName}(client, request),`,
      `    ...options,`,
      `  };`,
      `}`,
      ``,
    );

    // Mutation hook
    lines.push(
      `export function ${mutHookName}(`,
      `  client: RuntimeClient,`,
      `  options?: Partial<CreateMutationOptions<${responseType}, unknown, ${requestType}>>,`,
      `  queryClient?: QueryClient,`,
      `): CreateMutationResult<${responseType}, unknown, ${requestType}> {`,
      `  const mutationOptions = ${mutOptsFnName}(client, options);`,
      `  return createMutation(mutationOptions, queryClient);`,
      `}`,
      ``,
    );
  }

  return lines.join("\n");
}

function generateIndex(serviceNames: string[]): string {
  const lines = [`// Generated by generate-query-hooks.ts — DO NOT EDIT`, ``];
  for (const name of serviceNames) {
    lines.push(`export * from "./${toKebabCase(name)}";`);
  }
  lines.push(``);
  return lines.join("\n");
}

// --- Main ---

const outDir = path.resolve(__dirname, "gen");

fs.mkdirSync(outDir, { recursive: true });

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
  const code = generateServiceFile(descriptor);
  const fileName = `${toKebabCase(name)}.ts`;
  const filePath = path.join(outDir, fileName);
  fs.writeFileSync(filePath, code);

  const methods = extractMethods(descriptor);
  const queries = methods.filter((m) => m.classification === "query").length;
  const mutations = methods.filter(
    (m) => m.classification === "mutation",
  ).length;
  const skipped = Object.keys(descriptor.methods).length - methods.length;
  totalQueries += queries;
  totalMutations += mutations;

  console.log(
    `  ${fileName}: ${queries} queries, ${mutations} mutations, ${skipped} skipped`,
  );
}

// Generate barrel index
const indexCode = generateIndex(services.map((s) => s.name));
fs.writeFileSync(path.join(outDir, "index.ts"), indexCode);

console.log(`\nTotal: ${totalQueries} queries, ${totalMutations} mutations`);
console.log(`Output: ${outDir}`);
