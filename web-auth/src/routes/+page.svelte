<script>
  import Auth from "../components/Auth.svelte";

  const cloudClientIDs = import.meta.env.VITE_RILL_CLOUD_AUTH0_CLIENT_IDS;
  const disableForgotPassDomains = import.meta.env
    .VITE_DISABLE_FORGOT_PASS_DOMAINS;
  const connectionMap = import.meta.env.VITE_CONNECTION_MAP;
</script>

<Auth
  {cloudClientIDs}
  {disableForgotPassDomains}
  {connectionMap}
  configParams="@@config@@"
/>
