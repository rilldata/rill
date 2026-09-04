<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import FeedbackInbox from "@rilldata/web-common/features/chat/feedback-inbox/FeedbackInbox.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  const { feedbackInbox } = featureFlags;
</script>

<svelte:head>
  <title>{m.feedback_inbox_title()}</title>
</svelte:head>

{#if $feedbackInbox}
  <ContentContainer title={m.feedback_inbox_title()}>
    <FeedbackInbox />
  </ContentContainer>
{/if}
