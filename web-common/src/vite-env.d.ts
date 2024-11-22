// This file lets Typescript know about our custom environment variables
// See: https://vite.dev/guide/env-and-mode#intellisense-for-typescript

/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly RILL_UI_PUBLIC_POSTHOG_API_KEY: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
