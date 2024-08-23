<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { CANONICAL_ADMIN_URL } from "../../client/http-client";

  function handleSignIn() {
    window.location.href = `${CANONICAL_ADMIN_URL}/auth/login?redirect=${window.location.origin}${window.location.pathname}`;
  }
</script>

<Button type="primary" on:click={handleSignIn}>Log In / Sign Up</Button>
