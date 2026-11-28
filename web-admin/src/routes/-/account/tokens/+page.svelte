<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import type { V1UserAuthToken } from "@rilldata/web-admin/client";
  import CreateTokenDialog from "@rilldata/web-admin/features/personal-access-tokens/CreateTokenDialog.svelte";
  import PersonalAccessTokensTable from "@rilldata/web-admin/features/personal-access-tokens/PersonalAccessTokensTable.svelte";
  import RevokeTokenDialog from "@rilldata/web-admin/features/personal-access-tokens/RevokeTokenDialog.svelte";
  import { createPersonalAccessTokensQuery } from "@rilldata/web-admin/features/personal-access-tokens/personal-access-tokens";
  import { Button } from "@rilldata/web-common/components/button";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ArrowLeftIcon, KeyRoundIcon, PlusIcon } from "lucide-svelte";

  const tokensQuery = createPersonalAccessTokensQuery();

  // Back returns to the page the account page was opened from, or home when it was opened directly.
  let backHref = $state("/");
  afterNavigate(({ from }) => {
    // On a full page load `from` has no URL.
    if (from?.url && !from.route.id?.startsWith("/-/account")) {
      backHref = from.url.pathname + from.url.search;
    }
  });

  let tokens = $derived($tokensQuery.data ?? []);

  let createOpen = $state(false);
  // The revoke dialog is rendered outside the table:
  // revoking refreshes the list, which would unmount a dialog rendered in a row.
  // It is remounted for every token, so an error from an earlier attempt doesn't carry over.
  let revokingToken = $state<V1UserAuthToken | null>(null);
  let revokeOpen = $state(false);
  let revokeCount = $state(0);
</script>

<svelte:head>
  <title>{m.access_tokens_page_head_title()}</title>
</svelte:head>

<ContentContainer maxWidth={1100}>
  <a
    href={backHref}
    class="flex w-fit items-center gap-x-1 text-sm text-fg-secondary hover:text-fg-primary"
  >
    <ArrowLeftIcon size="14px" />
    {m.common_back()}
  </a>
  <h1 class="text-2xl font-semibold text-fg-primary">
    {m.access_tokens_page_title()}
  </h1>
  <section class="flex flex-col gap-y-4 pt-4">
    <div class="flex items-start gap-x-6">
      <div class="flex flex-col gap-y-1">
        <h2 class="text-lg font-semibold text-fg-primary">
          {m.access_tokens_title()}
        </h2>
        <p class="text-sm text-fg-secondary max-w-3xl">
          {m.access_tokens_description()}
          <a
            href="https://docs.rilldata.com/guide/administration/access-tokens/user-tokens"
            target="_blank"
            rel="noopener noreferrer"
          >
            {m.common_learn_more()}
          </a>
        </p>
      </div>
      {#if tokens.length > 0}
        <Button
          type="primary"
          class="ml-auto flex-none"
          onClick={() => (createOpen = true)}
        >
          <PlusIcon size="14px" />
          {m.access_tokens_new_token()}
        </Button>
      {/if}
    </div>

    {#if $tokensQuery.isPending}
      <div class="m-auto mt-20">
        <DelayedSpinner isLoading size="24px" />
      </div>
    {:else if $tokensQuery.isError}
      <div class="flex flex-col items-start gap-y-2">
        <p class="text-destructive text-sm">
          {m.access_tokens_load_failed({ message: $tokensQuery.error.message })}
        </p>
        <Button type="secondary" onClick={() => void $tokensQuery.refetch()}>
          {m.common_retry()}
        </Button>
      </div>
    {:else if tokens.length === 0}
      <div
        class="flex flex-col items-center gap-y-3 rounded-sm border border-dashed px-6 py-12 text-center"
      >
        <KeyRoundIcon size="28px" class="text-fg-muted" />
        <div class="flex flex-col gap-y-1">
          <span class="text-sm font-semibold text-fg-primary">
            {m.access_tokens_empty_title()}
          </span>
          <span class="text-sm text-fg-secondary max-w-md">
            {m.access_tokens_empty_description()}
          </span>
        </div>
        <Button type="primary" onClick={() => (createOpen = true)}>
          <PlusIcon size="14px" />
          {m.access_tokens_new_token()}
        </Button>
      </div>
    {:else}
      <PersonalAccessTokensTable
        {tokens}
        onRevoke={(token) => {
          revokingToken = token;
          revokeCount++;
          revokeOpen = true;
        }}
      />
    {/if}
  </section>
</ContentContainer>

<CreateTokenDialog bind:open={createOpen} />

{#if revokingToken}
  {#key revokeCount}
    <RevokeTokenDialog bind:open={revokeOpen} token={revokingToken} />
  {/key}
{/if}
