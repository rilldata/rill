import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
import { getContext, hasContext, setContext } from "svelte";

/**
 * The external AI clients we surface as one-click entry points. "other" opens the
 * generic connect flow without preselecting a provider.
 */
export type ConnectClientProvider = "claude" | "openai" | "gemini" | "other";

export interface ConnectClientContext {
  /**
   * Whether a connect flow is wired on the current surface. Presentational CTAs
   * self-hide when this is false, so surfaces that never provide the context
   * (e.g. Rill Developer, embedded dashboards) render nothing.
   */
  enabled: boolean;
  /** Open the connect dialog, optionally preselecting a provider. */
  open: (provider?: ConnectClientProvider) => void;
}

const CONNECT_CLIENT_CONTEXT_KEY = Symbol("connect-client-context");

/** Disabled default so the CTAs are inert unless a shell explicitly wires them. */
const DISABLED_CONTEXT: ConnectClientContext = {
  enabled: false,
  open: () => {},
};

export function setConnectClientContext(context: ConnectClientContext): void {
  setContext(CONNECT_CLIENT_CONTEXT_KEY, context);
}

export function getConnectClientContext(): ConnectClientContext {
  // Never surface connect CTAs inside an embedded dashboard, regardless of
  // whether a shell wired the context.
  if (EmbedStore.isEmbedded()) return DISABLED_CONTEXT;
  if (!hasContext(CONNECT_CLIENT_CONTEXT_KEY)) return DISABLED_CONTEXT;
  return getContext<ConnectClientContext>(CONNECT_CLIENT_CONTEXT_KEY);
}
