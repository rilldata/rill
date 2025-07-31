<script lang="ts">
  import { redirectToLoginFromComponent } from "@rilldata/web-admin/client/redirect-utils";
  import { Button } from "@rilldata/web-common/components/button";
</script>

<Button type="primary" onClick={redirectToLoginFromComponent}
  >Log In / Sign Up</Button
>
