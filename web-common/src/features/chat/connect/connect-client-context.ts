import { getContext, setContext } from "svelte";

export interface ConnectClientContext {
  /** Open the "connect your AI client" dialog. */
  open: () => void;
}

const CONNECT_CLIENT_CONTEXT_KEY = Symbol("connect-client-context");

export function setConnectClientContext(context: ConnectClientContext): void {
  setContext(CONNECT_CLIENT_CONTEXT_KEY, context);
}

/**
 * Returns the connect-client context, throwing if no provider is an ancestor.
 * The context is set only on Rill Cloud's AI chat page layout, which owns the
 * MCPConnectDialog; callers must gate on that (via the `adminServer` flag) before
 * calling this, so a missing provider surfaces as a real bug rather than a
 * silently hidden CTA.
 */
export function getConnectClientContext(): ConnectClientContext {
  const context = getContext<ConnectClientContext | undefined>(
    CONNECT_CLIENT_CONTEXT_KEY,
  );
  if (!context) {
    throw new Error(
      "getConnectClientContext() requires an ancestor that calls setConnectClientContext().",
    );
  }
  return context;
}
