<script lang="ts">
  import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
  import { Button } from "@rilldata/web-common/components/button";
</script>

<Button type="primary" onClick={redirectToLogin}>Log In / Sign Up</Button>
