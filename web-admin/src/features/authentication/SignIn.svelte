<script lang="ts">
  import { redirectToLoginFromComponent } from "@rilldata/web-admin/client/redirect-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
</script>

<Button type="primary" onClick={redirectToLoginFromComponent}
  >{m.auth_log_in_sign_up()}</Button
>
