<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { onMount } from "svelte";
  import CodeBlockInline from "../../../../../components/calls-to-action/CodeBlockInline.svelte";
  import CtaContentContainer from "../../../../../components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "../../../../../components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "../../../../../components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "../../../../../components/calls-to-action/CTAMessage.svelte";
  import KeyboardKey from "../../../../../components/calls-to-action/KeyboardKey.svelte";
  import GithubRepoInline from "../../../../../components/projects/GithubRepoInline.svelte";

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
    <CtaContentContainer>
      <Github className="w-10 h-10 text-gray-900" />
      <CtaHeader>Connect to Github</CtaHeader>
      <CtaMessage>
        You requested access to <GithubRepoInline githubUrl={remote} />. You can
        close this page now.
      </CtaMessage>
      <CtaMessage>
        The CLI will keep polling until Github access has been granted by an
        admin. You can stop polling by pressing <KeyboardKey label="Control" /> +
        <KeyboardKey label="C" /> and run <CodeBlockInline code="rill deploy" />
        again once access has been granted.
      </CtaMessage>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
