<script>
  import Auth from "../components/Auth.svelte";

  const cloudClientIDs = import.meta.env.VITE_RILL_CLOUD_AUTH0_CLIENT_IDS;
  const disableForgotPassDomains = import.meta.env
    .VITE_DISABLE_FORGOT_PASS_DOMAINS;
</script>

<Auth {cloudClientIDs} {disableForgotPassDomains} configParams="@@config@@" />
