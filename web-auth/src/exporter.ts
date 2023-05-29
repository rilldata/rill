import Auth from "./components/Auth.svelte";

const app = new Auth({
  target: document.body,
  props: {
    // This gets populated by Auth0 runtime
    configParams: "@@config@@",
    // This gets populated by RollUp and env variables
    cloudClientIDs: "%%VITE_RILL_CLOUD_AUTH0_CLIENT_IDS%%",
    disableForgotPassDomains: "%%VITE_DISABLE_FORGOT_PASS_DOMAINS%%",
    connectionMap: "%%VITE_CONNECTION_MAP%%",
  },
});

export default app;
