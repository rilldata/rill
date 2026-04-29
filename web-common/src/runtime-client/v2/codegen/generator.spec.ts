import { describe, it, expect } from "vitest";
import { MethodKind } from "@bufbuild/protobuf";
import {
  generateServiceFile,
  generateIndex,
  extractMethods,
  generateRawFunction,
  generateQueryMethod,
  generateInfiniteQueryMethod,
  generateMutationMethod,
  methodNames,
  requestTypes,
  type ServiceDef,
  type MethodContext,
} from "./generator";

// --- Mock service descriptors ---

class FakeRequestWithInstanceId {
  static typeName = "rill.runtime.v1.FakeRequestWithInstanceId";
  instanceId = "";
}

class FakeRequestWithoutInstanceId {
  static typeName = "rill.runtime.v1.FakeRequestWithoutInstanceId";
}

class FakeResponse {
  static typeName = "rill.runtime.v1.FakeResponse";
}

class FakePaginatedRequest {
  static typeName = "rill.runtime.v1.FakePaginatedRequest";
  instanceId = "";
  pageToken = "";
}

class FakePaginatedResponse {
  static typeName = "rill.runtime.v1.FakePaginatedResponse";
  nextPageToken = "";
}

/**
 * A minimal service with one query (with instanceId), one mutation
 * (without instanceId), one streaming method (should be skipped),
 * and one paginated query (generates infinite query hooks).
 */
const mockService: ServiceDef = {
  typeName: "rill.runtime.v1.RuntimeService",
  methods: {
    getFoo: {
      name: "GetFoo",
      I: FakeRequestWithInstanceId,
      O: FakeResponse,
      kind: MethodKind.Unary,
    },
    putBar: {
      name: "PutBar",
      I: FakeRequestWithoutInstanceId,
      O: FakeResponse,
      kind: MethodKind.Unary,
    },
    watchBaz: {
      name: "WatchBaz",
      I: FakeRequestWithInstanceId,
      O: FakeResponse,
      kind: MethodKind.ServerStreaming,
    },
    listItems: {
      name: "ListItems",
      I: FakePaginatedRequest,
      O: FakePaginatedResponse,
      kind: MethodKind.Unary,
    },
  },
};

// Shared context factory for narrow tests
function makeCtx(overrides: Partial<MethodContext> = {}): MethodContext {
  return {
    serviceName: "RuntimeService",
    serviceClientProp: "runtimeService",
    orvalTypes: new Set(),
    m: {
      methodKey: "getFoo",
      methodName: "GetFoo",
      inputType: "FakeRequestWithInstanceId",
      outputType: "FakeResponse",
      classification: "query",
      hasInstanceId: true,
      hasPageToken: false,
      hasNextPageToken: false,
    },
    ...overrides,
  };
}

// --- extractMethods ---

describe("extractMethods", () => {
  it("returns MethodInfo for unary methods only", () => {
    const methods = extractMethods(mockService);
    const keys = methods.map((m) => m.methodKey);
    expect(keys).toContain("getFoo");
    expect(keys).toContain("putBar");
    expect(keys).toContain("listItems");
    expect(keys).not.toContain("watchBaz");
  });

  it("classifies methods correctly", () => {
    const methods = extractMethods(mockService);
    const getFoo = methods.find((m) => m.methodKey === "getFoo");
    const putBar = methods.find((m) => m.methodKey === "putBar");
    expect(getFoo?.classification).toBe("query");
    expect(putBar?.classification).toBe("mutation");
  });

  it("detects hasInstanceId", () => {
    const methods = extractMethods(mockService);
    const getFoo = methods.find((m) => m.methodKey === "getFoo");
    const putBar = methods.find((m) => m.methodKey === "putBar");
    expect(getFoo?.hasInstanceId).toBe(true);
    expect(putBar?.hasInstanceId).toBe(false);
  });

  it("detects pagination fields", () => {
    const methods = extractMethods(mockService);
    const listItems = methods.find((m) => m.methodKey === "listItems");
    expect(listItems?.hasPageToken).toBe(true);
    expect(listItems?.hasNextPageToken).toBe(true);

    const getFoo = methods.find((m) => m.methodKey === "getFoo");
    expect(getFoo?.hasPageToken).toBe(false);
    expect(getFoo?.hasNextPageToken).toBe(false);
  });

  it("extracts short type names", () => {
    const methods = extractMethods(mockService);
    const getFoo = methods.find((m) => m.methodKey === "getFoo");
    expect(getFoo?.inputType).toBe("FakeRequestWithInstanceId");
    expect(getFoo?.outputType).toBe("FakeResponse");
    expect(getFoo?.methodName).toBe("GetFoo");
  });
});

// --- Per-method generators (narrow tests) ---

describe("methodNames", () => {
  it("generates correct name variants", () => {
    const names = methodNames(makeCtx());
    expect(names.rawFn).toBe("runtimeServiceGetFoo");
    expect(names.keyFn).toBe("getRuntimeServiceGetFooQueryKey");
    expect(names.optsFn).toBe("getRuntimeServiceGetFooQueryOptions");
    expect(names.hook).toBe("createRuntimeServiceGetFoo");
    expect(names.infOptsFn).toBe("getRuntimeServiceGetFooInfiniteQueryOptions");
    expect(names.infHook).toBe("createRuntimeServiceGetFooInfinite");
    expect(names.mutOptsFn).toBe("getRuntimeServiceGetFooMutationOptions");
    expect(names.mutHook).toBe("createRuntimeServiceGetFooMutation");
  });
});

describe("requestTypes", () => {
  it("wraps instanceId types in Omit", () => {
    const { requestType, requestSpread } = requestTypes(makeCtx());
    expect(requestType).toContain("Omit<");
    expect(requestType).toContain('"instanceId"');
    expect(requestSpread).toContain("instanceId: client.instanceId");
  });

  it("uses plain type when no instanceId", () => {
    const ctx = makeCtx({
      m: {
        ...makeCtx().m,
        hasInstanceId: false,
        inputType: "FakeRequestWithoutInstanceId",
      },
    });
    const { requestType, requestSpread } = requestTypes(ctx);
    expect(requestType).not.toContain("Omit");
    expect(requestSpread).toBe("request");
  });

  it("uses Orval V1 type when available", () => {
    const ctx = makeCtx({ orvalTypes: new Set(["V1FakeResponse"]) });
    const { responseType } = requestTypes(ctx);
    expect(responseType).toBe("V1FakeResponse");
  });

  it("falls back to PartialMessage when no Orval type", () => {
    const { responseType } = requestTypes(makeCtx());
    expect(responseType).toContain("PartialMessage<FakeResponse>");
  });
});

describe("generateRawFunction", () => {
  it("generates an async function with signal support", () => {
    const lines = generateRawFunction(makeCtx());
    const code = lines.join("\n");
    expect(code).toContain("export async function runtimeServiceGetFoo(");
    expect(code).toContain("signal?: AbortSignal");
    expect(code).toContain("fromJson(stripUndefined(");
    expect(code).toContain("toJson({ emitDefaultValues: true })");
  });
});

describe("generateQueryMethod", () => {
  it("generates query key, options, and hook", () => {
    const lines = generateQueryMethod(makeCtx());
    const code = lines.join("\n");
    expect(code).toContain("getRuntimeServiceGetFooQueryKey(");
    expect(code).toContain("getRuntimeServiceGetFooQueryOptions<");
    expect(code).toContain("createRuntimeServiceGetFoo<");
    expect(code).toContain("createQuery(");
  });

  it("does not generate a raw function (handled separately)", () => {
    const code = generateQueryMethod(makeCtx()).join("\n");
    expect(code).not.toContain("export async function");
  });
});

describe("generateInfiniteQueryMethod", () => {
  const paginatedCtx = makeCtx({
    m: {
      methodKey: "listItems",
      methodName: "ListItems",
      inputType: "FakePaginatedRequest",
      outputType: "FakePaginatedResponse",
      classification: "query",
      hasInstanceId: true,
      hasPageToken: true,
      hasNextPageToken: true,
    },
  });

  it("generates infinite query options and hook", () => {
    const code = generateInfiniteQueryMethod(paginatedCtx).join("\n");
    expect(code).toContain("getRuntimeServiceListItemsInfiniteQueryOptions<");
    expect(code).toContain("createRuntimeServiceListItemsInfinite<");
    expect(code).toContain("createInfiniteQuery(");
  });

  it("merges instanceId and pageToken into a single Omit", () => {
    const code = generateInfiniteQueryMethod(paginatedCtx).join("\n");
    expect(code).toContain('"instanceId" | "pageToken"');
    expect(code).not.toContain("Omit<Omit<");
  });

  it("uses pageToken as pageParam in queryFn", () => {
    const code = generateInfiniteQueryMethod(paginatedCtx).join("\n");
    expect(code).toContain("pageToken: pageParam");
  });

  it("extracts nextPageToken in getNextPageParam", () => {
    const code = generateInfiniteQueryMethod(paginatedCtx).join("\n");
    expect(code).toContain("nextPageToken");
    expect(code).toContain("getNextPageParam");
  });

  it("omits only pageToken when no instanceId", () => {
    const ctx = makeCtx({
      m: { ...paginatedCtx.m, hasInstanceId: false },
    });
    const code = generateInfiniteQueryMethod(ctx).join("\n");
    expect(code).toContain('"pageToken"');
    expect(code).not.toContain('"instanceId"');
  });
});

describe("generateMutationMethod", () => {
  const mutCtx = makeCtx({
    m: {
      methodKey: "putBar",
      methodName: "PutBar",
      inputType: "FakeRequestWithoutInstanceId",
      outputType: "FakeResponse",
      classification: "mutation",
      hasInstanceId: false,
      hasPageToken: false,
      hasNextPageToken: false,
    },
  });

  it("generates mutation options and hook", () => {
    const code = generateMutationMethod(mutCtx).join("\n");
    expect(code).toContain("getRuntimeServicePutBarMutationOptions(");
    expect(code).toContain("createRuntimeServicePutBarMutation(");
    expect(code).toContain("createMutation(");
  });

  it("does not generate a raw function (handled separately)", () => {
    const code = generateMutationMethod(mutCtx).join("\n");
    expect(code).not.toContain("export async function");
  });

  it("references the raw function in mutationFn", () => {
    const code = generateMutationMethod(mutCtx).join("\n");
    expect(code).toContain("runtimeServicePutBar(client, request)");
  });
});

// --- generateServiceFile ---

describe("generateServiceFile", () => {
  const { code: output } = generateServiceFile(mockService, new Set());

  it("starts with the DO NOT EDIT header", () => {
    expect(output).toMatch(/^\/\/ Generated by codegen\/run\.ts/);
  });

  it("imports stripUndefined", () => {
    expect(output).toContain(
      'import { stripUndefined } from "../strip-undefined"',
    );
  });

  it("returns extracted methods", () => {
    const { methods } = generateServiceFile(mockService, new Set());
    expect(methods.length).toBe(3); // getFoo, putBar, listItems (watchBaz skipped)
  });

  describe("query method (getFoo)", () => {
    it("generates 4 tiers: raw function, query key, query options, hook", () => {
      expect(output).toContain("export async function runtimeServiceGetFoo(");
      expect(output).toContain(
        "export function getRuntimeServiceGetFooQueryKey(",
      );
      expect(output).toContain(
        "export function getRuntimeServiceGetFooQueryOptions<",
      );
      expect(output).toContain("export function createRuntimeServiceGetFoo<");
    });

    it("uses Omit for instanceId request type", () => {
      expect(output).toMatch(
        /runtimeServiceGetFoo\(\s*client: RuntimeClient,\s*request: Omit</,
      );
    });

    it("injects instanceId in the request spread", () => {
      expect(output).toContain("instanceId: client.instanceId, ...request");
    });
  });

  describe("paginated query (listItems)", () => {
    it("generates both regular query and infinite query hooks", () => {
      expect(output).toContain("createRuntimeServiceListItems<");
      expect(output).toContain("createRuntimeServiceListItemsInfinite<");
    });

    it("imports infinite query types", () => {
      expect(output).toContain("createInfiniteQuery");
      expect(output).toContain("InfiniteData");
    });
  });

  describe("mutation method (putBar)", () => {
    it("generates 3 tiers: raw function, mutation options, mutation hook", () => {
      expect(output).toContain("export async function runtimeServicePutBar(");
      expect(output).toContain(
        "export function getRuntimeServicePutBarMutationOptions(",
      );
      expect(output).toContain(
        "export function createRuntimeServicePutBarMutation(",
      );
    });

    it("does not generate query key or query options", () => {
      expect(output).not.toContain("getRuntimeServicePutBarQueryKey");
      expect(output).not.toContain("getRuntimeServicePutBarQueryOptions");
    });

    it("does not use Omit (no instanceId)", () => {
      const putBarSection = output.slice(
        output.indexOf("runtimeServicePutBar("),
        output.indexOf("getRuntimeServicePutBarMutationOptions"),
      );
      expect(putBarSection).not.toContain("Omit");
    });
  });

  describe("streaming method (watchBaz)", () => {
    it("is not present in the output", () => {
      expect(output).not.toContain("watchBaz");
      expect(output).not.toContain("WatchBaz");
    });
  });

  describe("conditional imports", () => {
    it("omits mutation imports for query-only services", () => {
      const queryOnlyService: ServiceDef = {
        typeName: "rill.runtime.v1.ConnectorService",
        methods: {
          getFoo: mockService.methods.getFoo,
        },
      };
      const { code } = generateServiceFile(queryOnlyService, new Set());
      expect(code).not.toContain("createMutation");
      expect(code).not.toContain("CreateMutationOptions");
    });

    it("imports proto response types when no Orval counterpart exists", () => {
      // With no Orval types, response types must be imported from _pb.ts
      // so PartialMessage<FakeResponse> resolves in TypeScript
      expect(output).toContain("FakeResponse,");
      expect(output).toContain("FakePaginatedResponse,");
      expect(output).toContain("PartialMessage");
    });

    it("does not import proto response types when Orval counterpart exists", () => {
      const orvalTypes = new Set([
        "V1FakeRequestWithInstanceId",
        "V1FakeRequestWithoutInstanceId",
        "V1FakeResponse",
        "V1FakePaginatedRequest",
        "V1FakePaginatedResponse",
      ]);
      const { code } = generateServiceFile(mockService, orvalTypes);
      // Response types should come from Orval, not proto
      expect(code).toContain("V1FakeResponse");
      expect(code).not.toContain("PartialMessage");
    });
  });
});

// --- generateIndex ---

describe("generateIndex", () => {
  it("generates barrel exports in kebab-case", () => {
    const output = generateIndex(["QueryService", "RuntimeService"]);
    expect(output).toContain('export * from "./query-service"');
    expect(output).toContain('export * from "./runtime-service"');
  });

  it("includes the DO NOT EDIT header", () => {
    const output = generateIndex(["QueryService"]);
    expect(output).toMatch(/^\/\/ Generated by codegen\/run\.ts/);
  });
});
