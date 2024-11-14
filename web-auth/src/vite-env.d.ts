/// <reference types="svelte" />
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_RILL_CLOUD_AUTH0_CLIENT_IDS: string;
  readonly VITE_DISABLE_FORGOT_PASS_DOMAINS: string;
  readonly VITE_CONNECTION_MAP: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
