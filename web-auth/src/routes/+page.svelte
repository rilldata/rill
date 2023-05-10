<script>
  import Auth from "../components/Auth.svelte";

  const cloudClientIDs = import.meta.env.VITE_RILL_CLOUD_AUTH0_CLIENT_IDS;
  const oktaName = import.meta.env.VITE_OKTA_CONNECTION;
  const pingFedName = import.meta.env.VITE_PINGFED_CONNECTION;
  const disableForgotPassDomains = import.meta.env
    .VITE_DISABLE_FORGOT_PASS_DOMAINS;
</script>

<Auth
  {cloudClientIDs}
  {oktaName}
  {pingFedName}
  {disableForgotPassDomains}
  configParams="@@config@@"
/>
