<!--
  The web-local feedback inbox: cloud-deployment items via the LocalService
  bridge, plus the local instance's own items. Cloud editing has its own
  composition in web-admin (edit route) built from the same sections.
  Hosts render this inside a ContentContainer, which provides the page shell.
-->
<script lang="ts">
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { FeedbackStatus } from "./inbox-data";
  import FeedbackInboxHeader from "./FeedbackInboxHeader.svelte";
  import LocalServiceFeedbackSection from "./LocalServiceFeedbackSection.svelte";
  import RuntimeFeedbackSection from "./RuntimeFeedbackSection.svelte";

  const client = useRuntimeClient();
  const { readOnly } = featureFlags;

  let status: FeedbackStatus = "open";
  let categoryFilter: string | null = null;

  // Suggest-fix writes project files, so it needs a writable local instance.
  $: canSuggestFix = !$readOnly;

  // Clicking a category tag filters both sections; clicking it again clears.
  function toggleCategory(category: string) {
    categoryFilter = categoryFilter === category ? null : category;
  }
</script>

<FeedbackInboxHeader
  {status}
  onStatusChange={(s) => (status = s)}
  {categoryFilter}
  onClearCategory={() => (categoryFilter = null)}
/>
<div class="flex flex-col gap-y-6 pt-2">
  <LocalServiceFeedbackSection
    {status}
    {canSuggestFix}
    {categoryFilter}
    onCategoryClick={toggleCategory}
  />
  <RuntimeFeedbackSection
    {client}
    title={m.feedback_inbox_local_section()}
    {status}
    {canSuggestFix}
    {categoryFilter}
    onCategoryClick={toggleCategory}
  />
</div>
