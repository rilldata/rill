<script lang="ts">
  import type { V1UserAuthToken } from "@rilldata/web-admin/client";
  import CreateTokenDialog from "@rilldata/web-admin/features/personal-access-tokens/CreateTokenDialog.svelte";
  import PersonalAccessTokensTable from "@rilldata/web-admin/features/personal-access-tokens/PersonalAccessTokensTable.svelte";
  import RevokeTokenDialog from "@rilldata/web-admin/features/personal-access-tokens/RevokeTokenDialog.svelte";
  import { createPersonalAccessTokensQuery } from "@rilldata/web-admin/features/personal-access-tokens/personal-access-tokens";
  import { Button } from "@rilldata/web-common/components/button";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { KeyRoundIcon, PlusIcon } from "lucide-svelte";

  const tokensQuery = createPersonalAccessTokensQuery();

  let tokens = $derived($tokensQuery.data ?? []);

  let createOpen = $state(false);
  // The revoke dialog is rendered outside the table:
  // revoking refreshes the list, which would unmount a dialog rendered in a row.
  let revokingToken = $state<V1UserAuthToken | null>(null);
  let revokeOpen = $state(false);
</script>

<svelte:head>
  <title>{m.access_tokens_page_head_title()}</title>
</svelte:head>

<ContentContainer title={m.access_tokens_page_title()} maxWidth={1100}>
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
      <p class="text-destructive text-sm">
        {m.access_tokens_load_failed({ message: $tokensQuery.error.message })}
      </p>
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
          revokeOpen = true;
        }}
      />
    {/if}
  </section>
</ContentContainer>

<CreateTokenDialog bind:open={createOpen} />

{#if revokingToken}
  <RevokeTokenDialog bind:open={revokeOpen} token={revokingToken} />
{/if}
