<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import CtaHeader from "../../../../../components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "../../../../../components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "../../../../../components/calls-to-action/CTAMessage.svelte";

  const remote = new URLSearchParams(window.location.search).get("remote");
  const user = createAdminServiceGetCurrentUser({
    query: {
      onSuccess: (data) => {
        if (!data.user) {
          goto(`${ADMIN_URL}/auth/login?redirect=${window.location.href}`);
        }
      },
    },
  });
</script>

<svelte:head>
  <title>Github access requested</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <CtaLayoutContainer>
    <CtaHeader>Connect to Github</CtaHeader>
    <CtaMessage>
      You requested access to {@html remote}. You can close this page now.<br />
      CLI will keep polling until access has been granted by admin.<br />
      You can stop polling by pressing `ctrl+c` and run `rill deploy` again once
      access has been granted.
    </CtaMessage>
  </CtaLayoutContainer>
{/if}
