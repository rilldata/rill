type RPCRequest = {
    id: string;
    method: string;
    params?: unknown;
};

type RPCResponse = {
    id: string;
    result?: unknown;
    error?: string;
};

type RPCMethods = {
    [key: string]: (params?: unknown) => Promise<unknown> | unknown;
};

const methods: RPCMethods = {
    echo(message: { message: string }) {
        return message;
    },
    setTheme({ theme }: { theme: string }) {
        document.documentElement.setAttribute("data-theme", theme);
        return { success: true };
    }
};

async function handleRPCMessage(event: MessageEvent<RPCRequest>) {

    const { id, method, params } = event.data;

    if (methods[method]) {
        try {
            const result = await methods[method](params);
            if (id) {
                event.source?.postMessage({ id, result } as RPCResponse);
            }
        } catch (error) {
            if (id) {
                event.source?.postMessage({ id, error: (error as Error).message } as RPCResponse);
            }
        }
    }
}

export function initRPC() {
    window.addEventListener("message", (event) => {
        void handleRPCMessage(event);
    });
}

export function registerMethod<T>(name: string, func: (params: T) => Promise<unknown> | unknown) {
    methods[name] = func;
}

export function emit(method: string, params?: unknown) {
    if (window.parent !== window) {
        window.parent.postMessage({ method, params } as RPCRequest, "*");
    }
}
