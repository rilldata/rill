import Auth from "./components/Auth.svelte";

const app = new Auth({
  target: document.body,
  props: {
    cloudClientIDs: import.meta.env.VITE_RILL_CLOUD_AUTH0_CLIENT_IDS,
    disableForgotPassDomains: import.meta.env.VITE_DISABLE_FORGOT_PASS_DOMAINS,
    connectionMap: import.meta.env.VITE_CONNECTION_MAP,
    // This gets populated by Auth0 runtime
    configParams: "@@config@@",
  },
});

export default app;
