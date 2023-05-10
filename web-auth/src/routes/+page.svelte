<script>
  import Auth from "../components/Auth.svelte";

  let cloudClientIDs = import.meta.env.VITE_RILL_CLOUD_AUTH0_CLIENT_IDS;
</script>

<Auth {cloudClientIDs} configParams="@@config@@" />
