type JSONRPCRequest = {
  jsonrpc: "2.0";
  id: string | number | null;
  method: string;
  params?: unknown;
};

type JSONRPCResponse = {
  jsonrpc: "2.0";
  id: string | number | null;
  result?: unknown;
  error?: {
    code: number;
    message: string;
    data?: unknown;
  };
};

const JSONRPC_ERRORS = {
  PARSE_ERROR: { code: -32700, message: "Parse error" },
  INVALID_REQUEST: { code: -32600, message: "Invalid Request" },
  METHOD_NOT_FOUND: { code: -32601, message: "Method not found" },
  INVALID_PARAMS: { code: -32602, message: "Invalid params" },
  INTERNAL_ERROR: { code: -32603, message: "Internal error" },
};

type JSONRPCMethods = {
  [key: string]: (params?: unknown) => Promise<unknown> | unknown;
};

const methods: JSONRPCMethods = {
  echo(message: { message: string }) {
    return message;
  },
};

async function handleRPCMessage(event: MessageEvent<JSONRPCRequest>) {
  if (typeof event.data !== "object" || event.data === null) {
    return sendError(null, JSONRPC_ERRORS.INVALID_REQUEST);
  }

  const { id, method, params } = event.data;

  if (
    typeof method !== "string" ||
    (id !== null && typeof id !== "string" && typeof id !== "number")
  ) {
    return sendError(id, JSONRPC_ERRORS.INVALID_REQUEST);
  }

  if (!methods[method]) {
    return sendError(id, JSONRPC_ERRORS.METHOD_NOT_FOUND);
  }

  try {
    const result = await methods[method](params);
    if (id !== null) {
      sendResponse(id, result);
    }
  } catch (error) {
    sendError(id, {
      code: JSONRPC_ERRORS.INTERNAL_ERROR.code,
      message: (error as Error).message,
    });
  }
}

function sendResponse(id: string | number | null, result: unknown) {
  if (window.parent !== window) {
    window.parent.postMessage(
      { jsonrpc: "2.0", id, result } as JSONRPCResponse,
      "*",
    );
  }
}

function sendError(
  id: string | number | null,
  error: { code: number; message: string; data?: unknown },
) {
  if (window.parent !== window) {
    window.parent.postMessage(
      { jsonrpc: "2.0", id, error } as JSONRPCResponse,
      "*",
    );
  }
}

export function initRPC() {
  window.removeEventListener("message", (_event: MessageEvent) => {});
  window.addEventListener("message", (event: MessageEvent) => {
    if (event.source && event.data) {
      void handleRPCMessage(event as MessageEvent<JSONRPCRequest>);
    }
  });
}

export function registerMethod<T>(
  name: string,
  func: (params: T) => Promise<unknown> | unknown,
) {
  methods[name] = func;
}

export function emit(
  method: string,
  params?: unknown,
  id: string | number | null = null,
) {
  if (window.parent !== window) {
    window.parent.postMessage(
      { jsonrpc: "2.0", id, method, params } as JSONRPCRequest,
      "*",
    );
  }
}
