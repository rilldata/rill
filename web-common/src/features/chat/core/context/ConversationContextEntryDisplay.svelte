<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import {
    ContextTypeData,
    type ConversationContextEntry,
  } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { ConversationContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

  export let context: ConversationContext;
  export let entry: ConversationContextEntry;

  $: ({ instanceId } = $runtime);

  $: contextRecord = context.record;
  $: ({ type, value } = entry);

  $: data = ContextTypeData[type];
  $: ({ icon, formatter } = data);

  $: formattedValue = formatter(value, $contextRecord, instanceId);
</script>

<Chip>
  <svelte:fragment slot="body">
    <svelte:component this={icon} size="16px" />
    <span>{$formattedValue}</span>
  </svelte:fragment>
</Chip>
