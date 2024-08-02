<script lang="ts">
  import { createAdminServiceListProjectWhitelistedDomains } from "@rilldata/web-admin/client";

  export let organization: string;
  export let project: string;

  $: allowedDomains = createAdminServiceListProjectWhitelistedDomains(
    organization,
    project,
  );
</script>

<div class="text-xs text-gray-500">
  {#if $allowedDomains.data?.domains?.length}
    Anyone with a {#each $allowedDomains.data?.domains as { domain }, index (domain)}
      <b>@{domain}</b>{#if index < $allowedDomains.data?.domains?.length - 1}
        <span class="m-0.5">or</span>
      {/if}
    {/each}
    email address can join this project as a Viewer.
  {:else}
    You have not enabled access by domain.
  {/if}
  <a
    target="_blank"
    href="https://docs.rilldata.com/reference/cli/user/whitelist/"
  >
    Learn more
  </a>
</div>
