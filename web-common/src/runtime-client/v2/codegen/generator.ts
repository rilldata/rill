/**
 * Code generator library: reads ConnectRPC service descriptors and
 * produces TanStack Query hooks for Svelte.
 *
 * This module is a pure library with no side effects.
 * Run via: tsx src/runtime-client/v2/codegen/run.ts
 */

import * as fs from "node:fs";
import * as path from "node:path";
import { MethodKind } from "@bufbuild/protobuf";
import { classifyMethod, type MethodClassification } from "./config";

export interface ServiceDef {
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

export interface MethodInfo {
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
  /** Whether the request type has a pageToken field (pagination input) */
  hasPageToken: boolean;
  /** Whether the response type has a nextPageToken field (pagination output) */
  hasNextPageToken: boolean;
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
  if (!(serviceName in fileMap)) {
    throw new Error(
      `Unknown service "${serviceName}"; add it to the fileMap in getProtoImportPath`,
    );
  }
  return `../../../proto/gen/rill/runtime/v1/${fileMap[serviceName]}`;
}

// --- JSON bridge: Orval type helpers ---

const ORVAL_IMPORT_PATH = "../../gen/index.schemas";

/** Scan the Orval schema file to discover available V1 type names */
function loadOrvalTypes(baseDir: string): Set<string> {
  const schemaPath = path.resolve(baseDir, "../../gen/index.schemas.ts");
  const content = fs.readFileSync(schemaPath, "utf-8");
  const types = new Set<string>();
  for (const match of content.matchAll(/^export (?:type|interface) (\w+)/gm)) {
    types.add(match[1]);
  }
  return types;
}

function orvalTypeName(protoTypeName: string): string {
  return `V1${protoTypeName}`;
}

function hasOrvalType(
  availableOrvalTypes: Set<string>,
  protoTypeName: string,
): boolean {
  return availableOrvalTypes.has(orvalTypeName(protoTypeName));
}

/** Get the public-facing type for a request or response */
function publicType(
  availableOrvalTypes: Set<string>,
  protoTypeName: string,
): string {
  return hasOrvalType(availableOrvalTypes, protoTypeName)
    ? orvalTypeName(protoTypeName)
    : `PartialMessage<${protoTypeName}>`;
}

// --- Method extraction ---

function extractMethods(service: ServiceDef): MethodInfo[] {
  const serviceName = extractShortName(service.typeName);
  const methods: MethodInfo[] = [];

  for (const [key, method] of Object.entries(service.methods)) {
    if (method.kind !== MethodKind.Unary) {
      // Skip streaming methods automatically
      continue;
    }

    const classification = classifyMethod(serviceName, key);
    if (classification === "skip") continue;

    const requestInstance = new method.I();
    const hasInstanceId = "instanceId" in requestInstance;
    const hasPageToken = "pageToken" in requestInstance;

    // Check response for nextPageToken (pagination output)
    const ResponseType = method.O as unknown as {
      new (): Record<string, unknown>;
    };
    const hasNextPageToken = "nextPageToken" in new ResponseType();

    methods.push({
      methodKey: key,
      methodName: method.name,
      inputType: extractShortName(method.I.typeName),
      outputType: extractShortName(method.O.typeName),
      classification,
      hasInstanceId,
      hasPageToken,
      hasNextPageToken,
    });
  }

  return methods;
}

// --- Code generation: per-method helpers ---

interface FileContext {
  serviceName: string;
  serviceClientProp: string;
  orvalTypes: Set<string>;
}

interface MethodContext extends FileContext {
  m: MethodInfo;
}

function methodNames(ctx: MethodContext) {
  const { serviceClientProp, m } = ctx;
  const base = `${serviceClientProp}${pascalCase(m.methodKey)}`;
  return {
    rawFn: base,
    keyFn: `get${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}QueryKey`,
    optsFn: `get${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}QueryOptions`,
    hook: `create${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}`,
    infOptsFn: `get${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}InfiniteQueryOptions`,
    infHook: `create${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}Infinite`,
    mutOptsFn: `get${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}MutationOptions`,
    mutHook: `create${pascalCase(serviceClientProp)}${pascalCase(m.methodKey)}Mutation`,
  };
}

function requestTypes(ctx: MethodContext) {
  const { m, orvalTypes } = ctx;
  const inputPublic = publicType(orvalTypes, m.inputType);
  const requestType = m.hasInstanceId
    ? `Omit<${inputPublic}, "instanceId">`
    : inputPublic;
  const requestSpread = m.hasInstanceId
    ? `{ instanceId: client.instanceId, ...request }`
    : `request`;
  const responseType = publicType(orvalTypes, m.outputType);
  return { requestType, requestSpread, responseType };
}

function generateRawFunction(ctx: MethodContext): string[] {
  const { serviceName, serviceClientProp, m } = ctx;
  const { rawFn } = methodNames(ctx);
  const { requestType, requestSpread, responseType } = requestTypes(ctx);

  return [
    `/**`,
    ` * Raw RPC call: ${serviceName}.${m.methodName}`,
    ` */`,
    `export async function ${rawFn}(`,
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
  ];
}

function generateQueryMethod(ctx: MethodContext): string[] {
  const { serviceName, m } = ctx;
  const { rawFn, keyFn, optsFn, hook } = methodNames(ctx);
  const { requestType, responseType } = requestTypes(ctx);

  return [
    // Query key
    `export function ${keyFn}(`,
    `  instanceId: string,`,
    `  request?: ${requestType},`,
    `): QueryKey {`,
    `  return ["${serviceName}", "${m.methodKey}", instanceId, request ?? {}] as const;`,
    `}`,
    ``,
    // Query options
    `export function ${optsFn}<TData = ${responseType}>(`,
    `  client: RuntimeClient,`,
    `  request: ${requestType},`,
    `  options?: {`,
    `    query?: Partial<CreateQueryOptions<${responseType}, ConnectError, TData>>;`,
    `  },`,
    `): CreateQueryOptions<${responseType}, ConnectError, TData> & { queryKey: QueryKey } {`,
    `  const queryKey = ${keyFn}(client.instanceId, request);`,
    `  const queryFn: QueryFunction<${responseType}> = ({ signal }) =>`,
    `    ${rawFn}(client, request, { signal });`,
    `  return {`,
    `    queryKey,`,
    `    queryFn,`,
    `    enabled: !!client.instanceId,`,
    `    ...options?.query,`,
    `  } as CreateQueryOptions<${responseType}, ConnectError, TData> & { queryKey: QueryKey };`,
    `}`,
    ``,
    // Hook
    `export function ${hook}<TData = ${responseType}>(`,
    `  client: RuntimeClient,`,
    `  request: ${requestType},`,
    `  options?: {`,
    `    query?: Partial<CreateQueryOptions<${responseType}, ConnectError, TData>>;`,
    `  },`,
    `  queryClient?: QueryClient,`,
    `): CreateQueryResult<TData, ConnectError> {`,
    `  const queryOptions = ${optsFn}(client, request, options);`,
    `  return createQuery(queryOptions, queryClient);`,
    `}`,
    ``,
  ];
}

function generateInfiniteQueryMethod(ctx: MethodContext): string[] {
  const { m } = ctx;
  const { rawFn, keyFn, infOptsFn, infHook } = methodNames(ctx);
  const { requestType, responseType } = requestTypes(ctx);

  // For infinite queries, the request type omits pageToken (managed by TanStack).
  // Merge Omit keys when instanceId is already omitted to avoid Omit<Omit<...>> nesting.
  const omitKeys = m.hasInstanceId
    ? `"instanceId" | "pageToken"`
    : `"pageToken"`;
  const inputPublic = publicType(ctx.orvalTypes, m.inputType);
  const paginatedRequestType = `Omit<${inputPublic}, ${omitKeys}>`;

  return [
    // Infinite query options
    `export function ${infOptsFn}<TData = InfiniteData<${responseType}>>(`,
    `  client: RuntimeClient,`,
    `  request: ${paginatedRequestType},`,
    `  options?: {`,
    `    query?: Partial<CreateInfiniteQueryOptions<${responseType}, ConnectError, TData, ${responseType}, QueryKey, string | undefined>>;`,
    `  },`,
    `): CreateInfiniteQueryOptions<${responseType}, ConnectError, TData, ${responseType}, QueryKey, string | undefined> & { queryKey: QueryKey } {`,
    `  const queryKey = [...${keyFn}(client.instanceId, request), "infinite"] as QueryKey;`,
    `  return {`,
    `    queryKey,`,
    `    queryFn: ({ pageParam, signal }) =>`,
    `      ${rawFn}(client, { ...request, pageToken: pageParam } as ${requestType}, { signal }),`,
    `    initialPageParam: undefined as string | undefined,`,
    `    getNextPageParam: (lastPage) =>`,
    `      (lastPage as Record<string, unknown>)?.nextPageToken as string | undefined || undefined,`,
    `    enabled: !!client.instanceId,`,
    `    ...options?.query,`,
    `  } as CreateInfiniteQueryOptions<${responseType}, ConnectError, TData, ${responseType}, QueryKey, string | undefined> & { queryKey: QueryKey };`,
    `}`,
    ``,
    // Hook
    `export function ${infHook}<TData = InfiniteData<${responseType}>>(`,
    `  client: RuntimeClient,`,
    `  request: ${paginatedRequestType},`,
    `  options?: {`,
    `    query?: Partial<CreateInfiniteQueryOptions<${responseType}, ConnectError, TData, ${responseType}, QueryKey, string | undefined>>;`,
    `  },`,
    `  queryClient?: QueryClient,`,
    `): CreateInfiniteQueryResult<TData, ConnectError> {`,
    `  const queryOptions = ${infOptsFn}(client, request, options);`,
    `  return createInfiniteQuery(queryOptions, queryClient);`,
    `}`,
    ``,
  ];
}

function generateMutationMethod(ctx: MethodContext): string[] {
  const { rawFn, mutOptsFn, mutHook } = methodNames(ctx);
  const { requestType, responseType } = requestTypes(ctx);

  return [
    // Mutation options
    `export function ${mutOptsFn}(`,
    `  client: RuntimeClient,`,
    `  options?: Partial<CreateMutationOptions<${responseType}, unknown, ${requestType}>>,`,
    `): CreateMutationOptions<${responseType}, unknown, ${requestType}> {`,
    `  return {`,
    `    mutationFn: (request) => ${rawFn}(client, request),`,
    `    ...options,`,
    `  };`,
    `}`,
    ``,
    // Mutation hook
    `export function ${mutHook}(`,
    `  client: RuntimeClient,`,
    `  options?: Partial<CreateMutationOptions<${responseType}, unknown, ${requestType}>>,`,
    `  queryClient?: QueryClient,`,
    `): CreateMutationResult<${responseType}, unknown, ${requestType}> {`,
    `  const mutationOptions = ${mutOptsFn}(client, options);`,
    `  return createMutation(mutationOptions, queryClient);`,
    `}`,
    ``,
  ];
}

// --- Code generation: file assembly ---

interface GenerateResult {
  code: string;
  methods: MethodInfo[];
}

function generateServiceFile(
  service: ServiceDef,
  availableOrvalTypes: Set<string>,
): GenerateResult {
  const serviceName = extractShortName(service.typeName);
  const serviceClientProp = getServiceClientProp(serviceName);
  const protoImportPath = getProtoImportPath(serviceName);
  const methods = extractMethods(service);

  const queryMethods = methods.filter((m) => m.classification === "query");
  const mutationMethods = methods.filter(
    (m) => m.classification === "mutation",
  );
  const infiniteQueryMethods = queryMethods.filter(
    (m) => m.hasPageToken && m.hasNextPageToken,
  );

  // Collect proto types (value imports; needed for fromJson/toJson calls)
  const protoTypes = new Set<string>();
  // Collect Orval types (type-only imports; used in public API signatures)
  const orvalTypeImports = new Set<string>();
  // Track whether PartialMessage is needed (for request types without V1 counterparts)
  let needsPartialMessage = false;
  const needsInfiniteQuery = infiniteQueryMethods.length > 0;

  for (const m of methods) {
    protoTypes.add(m.inputType);

    if (hasOrvalType(availableOrvalTypes, m.inputType)) {
      orvalTypeImports.add(orvalTypeName(m.inputType));
    } else {
      needsPartialMessage = true;
    }

    if (hasOrvalType(availableOrvalTypes, m.outputType)) {
      orvalTypeImports.add(orvalTypeName(m.outputType));
    } else {
      protoTypes.add(m.outputType);
      needsPartialMessage = true;
    }
  }

  const lines: string[] = [];

  // Header
  lines.push(`// Generated by codegen/run.ts — DO NOT EDIT`);
  lines.push(``);

  // Imports are sorted to match VS Code's "Organize Imports" order:
  // package imports first (alphabetical), then relative imports (alphabetical).
  // Specifiers within each import: values first, then types, each alphabetical.

  const hasQueries = queryMethods.length > 0;
  const hasMutations = mutationMethods.length > 0;

  // --- Package imports (sorted by module path) ---

  // @bufbuild/protobuf
  const bufSpecs: string[] = ["JsonValue"];
  if (needsPartialMessage) bufSpecs.push("PartialMessage");
  bufSpecs.sort();
  lines.push(
    `import type { ${bufSpecs.join(", ")} } from "@bufbuild/protobuf";`,
  );

  // @connectrpc/connect
  lines.push(`import type { ConnectError } from "@connectrpc/connect";`);

  // @tanstack/svelte-query (mixed import: values first, then inline types)
  const tanstackValues: string[] = [];
  const tanstackTypes: string[] = [];
  if (hasQueries) {
    tanstackValues.push("createQuery");
    tanstackTypes.push(
      "CreateQueryOptions",
      "CreateQueryResult",
      "QueryFunction",
    );
  }
  if (hasMutations) {
    tanstackValues.push("createMutation");
    tanstackTypes.push("CreateMutationOptions", "CreateMutationResult");
  }
  if (hasQueries || hasMutations) {
    tanstackTypes.push("QueryClient", "QueryKey");
  }
  if (needsInfiniteQuery) {
    tanstackValues.push("createInfiniteQuery");
    tanstackTypes.push(
      "CreateInfiniteQueryOptions",
      "CreateInfiniteQueryResult",
      "InfiniteData",
    );
  }
  tanstackValues.sort();
  tanstackTypes.sort();
  if (tanstackValues.length > 0 || tanstackTypes.length > 0) {
    lines.push(
      `import {`,
      ...tanstackValues.map((v) => `  ${v},`),
      ...tanstackTypes.map((t) => `  type ${t},`),
      `} from "@tanstack/svelte-query";`,
    );
  }

  // --- Relative imports (sorted by module path) ---

  // Proto type imports (value; needed for fromJson calls)
  const sortedProto = [...protoTypes].sort();
  lines.push(
    `import {`,
    ...sortedProto.map((t) => `  ${t},`),
    `} from "${protoImportPath}";`,
  );

  // Orval type imports (type-only; for public API signatures)
  if (orvalTypeImports.size > 0) {
    const sortedOrval = [...orvalTypeImports].sort();
    lines.push(
      `import type {`,
      ...sortedOrval.map((t) => `  ${t},`),
      `} from "${ORVAL_IMPORT_PATH}";`,
    );
  }

  // RuntimeClient
  lines.push(`import type { RuntimeClient } from "../runtime-client";`);

  // stripUndefined (proto fromJson rejects undefined values;
  // Orval's HTTP client silently omitted them)
  lines.push(`import { stripUndefined } from "../strip-undefined";`);

  lines.push(``);

  const fileCtx: FileContext = {
    serviceName,
    serviceClientProp,
    orvalTypes: availableOrvalTypes,
  };

  for (const m of queryMethods) {
    const ctx: MethodContext = { ...fileCtx, m };
    lines.push(...generateRawFunction(ctx));
    lines.push(...generateQueryMethod(ctx));
  }

  for (const m of infiniteQueryMethods) {
    lines.push(...generateInfiniteQueryMethod({ ...fileCtx, m }));
  }

  for (const m of mutationMethods) {
    const ctx: MethodContext = { ...fileCtx, m };
    lines.push(...generateRawFunction(ctx));
    lines.push(...generateMutationMethod(ctx));
  }

  return { code: lines.join("\n"), methods };
}

function generateIndex(serviceNames: string[]): string {
  const lines = [`// Generated by codegen/run.ts — DO NOT EDIT`, ``];
  for (const name of serviceNames) {
    lines.push(`export * from "./${toKebabCase(name)}";`);
  }
  lines.push(``);
  return lines.join("\n");
}

export {
  generateServiceFile,
  generateIndex,
  extractMethods,
  loadOrvalTypes,
  toKebabCase,
  // Per-method generators (exported for narrow testing)
  generateRawFunction,
  generateQueryMethod,
  generateInfiniteQueryMethod,
  generateMutationMethod,
  methodNames,
  requestTypes,
  type FileContext,
  type MethodContext,
  type GenerateResult,
};
