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
 * Returns the connect-client context, or undefined if no provider is an ancestor.
 * The context is set only on Rill Cloud chat surfaces that own an MCPConnectDialog
 * (the AI chat page and dashboard/canvas layouts); chat surfaces without a
 * provider (Rill Developer, embeds, edit mode) render no connect CTA.
 */
export function getConnectClientContext(): ConnectClientContext | undefined {
  return getContext<ConnectClientContext | undefined>(
    CONNECT_CLIENT_CONTEXT_KEY,
  );
}
