<script lang="ts">
  import { page } from "$app/stores";
  import MemoryManager from "@rilldata/web-common/features/chat/memory/MemoryManager.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium text-fg-primary">{m.chat_memory_title()}</h2>
  <MemoryManager conversationBasePath={`/${organization}/${project}/-/ai`} />
</div>
