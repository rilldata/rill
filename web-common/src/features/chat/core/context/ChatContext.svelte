<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getContextOptions } from "@rilldata/web-common/features/chat/core/context/context-options.ts";
  import {
    type ChatContextEntry,
    ChatContextEntryType,
    type ContextMetadata,
    ContextTypeData,
  } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { builderActions, getAttrs } from "bits-ui";
  import { ChevronDownIcon } from "lucide-svelte";
  import { readable, writable } from "svelte/store";

  export let chatCtx: ChatContextEntry;
  export let metadata: ContextMetadata;
  export let onUpdate: () => void;

  const chatCtxStore = writable(chatCtx);
  $: chatCtxStore.set(chatCtx);

  const contextOptions = getContextOptions(chatCtxStore, readable(""));
  $: optionsForType = $contextOptions[chatCtx.type] ?? [];
  $: hasOptions = chatCtx.type in $contextOptions;

  export function getChatContext() {
    return chatCtx;
  }

  function updateValue(newValue: string) {
    if (chatCtx.type === ChatContextEntryType.DimensionValues) {
      chatCtx.subValue = newValue;
    } else {
      chatCtx.value = newValue;
    }
    chatCtx.label =
      ContextTypeData[chatCtx.type]?.getLabel(chatCtx, metadata) ?? newValue;

    setTimeout(onUpdate);
  }
</script>

<span class="inline-block gap-1 text-sm underline" contenteditable="false">
  <div class="flex flex-row items-center">
    <span>{chatCtx.label}</span>
    {#if hasOptions}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <button
            {...getAttrs([builder])}
            use:builderActions={{ builders: [builder] }}
            type="button"
          >
            <ChevronDownIcon size={12} />
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content side="top" sideOffset={8}>
          {#each optionsForType as option (option.value)}
            <DropdownMenu.Item on:click={() => updateValue(option.value)}>
              {option.label}
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  </div>
</span>
