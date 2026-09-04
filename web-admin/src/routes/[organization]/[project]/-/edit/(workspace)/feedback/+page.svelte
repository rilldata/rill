<!--
  AI feedback inbox for cloud edit sessions. Feedback from dashboard users lives
  in the production deployment's runtime store, so this page reads it there
  directly; a second section covers feedback submitted inside this edit session.
  Suggest-fix and file writes go to the edit session's runtime (ambient context),
  which has the repo checkout.
-->
<script lang="ts">
  import { page } from "$app/stores";
  import { createProdRuntimeClient } from "@rilldata/web-admin/features/ai-feedback/prod-runtime-client";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import FeedbackInboxHeader from "@rilldata/web-common/features/chat/feedback-inbox/FeedbackInboxHeader.svelte";
  import RuntimeFeedbackSection from "@rilldata/web-common/features/chat/feedback-inbox/RuntimeFeedbackSection.svelte";
  import type { FeedbackStatus } from "@rilldata/web-common/features/chat/feedback-inbox/inbox-data";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const { feedbackInbox } = featureFlags;

  // Ambient runtime = this edit session's deployment.
  const editClient = useRuntimeClient();

  $: ({ organization, project } = $page.params);

  // Feedback from dashboard users lives on the production deployment.
  $: prodRuntime = createProdRuntimeClient(organization, project);

  let status: FeedbackStatus = "open";
  let categoryFilter: string | null = null;

  // Clicking a category tag filters both sections; clicking it again clears.
  function toggleCategory(category: string) {
    categoryFilter = categoryFilter === category ? null : category;
  }
</script>

<svelte:head>
  <title>{m.feedback_inbox_title()}</title>
</svelte:head>

{#if $feedbackInbox}
  <ContentContainer title={m.feedback_inbox_title()}>
    <FeedbackInboxHeader
      {status}
      onStatusChange={(s) => (status = s)}
      {categoryFilter}
      onClearCategory={() => (categoryFilter = null)}
    />
    <div class="flex flex-col gap-y-6 pt-2">
      <RuntimeFeedbackSection
        client={$prodRuntime.client}
        title={m.feedback_inbox_prod_section()}
        {status}
        canSuggestFix={true}
        {categoryFilter}
        onCategoryClick={toggleCategory}
        unavailableMessage={$prodRuntime.resolved
          ? m.feedback_inbox_not_deployed()
          : ""}
      />
      <RuntimeFeedbackSection
        client={editClient}
        title={m.feedback_inbox_edit_session_section()}
        {status}
        canSuggestFix={true}
        {categoryFilter}
        onCategoryClick={toggleCategory}
      />
    </div>
  </ContentContainer>
{/if}
