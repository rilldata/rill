<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { ADMIN_URL } from "../../client/http-client";

  function handleSignIn() {
    window.location.href = `${ADMIN_URL}/auth/login?redirect=${window.origin}`;
  }
</script>

<Button type="primary" on:click={handleSignIn}>Log In / Sign Up</Button>
