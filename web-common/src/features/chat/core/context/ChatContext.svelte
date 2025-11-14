<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getContextOptions } from "@rilldata/web-common/features/chat/core/context/context-options.ts";
  import {
    type ChatContextEntry,
    type ContextMetadata,
    ContextTypeData,
  } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { convertContextToAttrs } from "@rilldata/web-common/features/chat/core/context/conversions.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import { ChevronDownIcon } from "lucide-svelte";
  import { readable } from "svelte/store";

  export let chatCtx: ChatContextEntry;
  export let metadata: ContextMetadata;
  export let onUpdate: () => void;

  const contextOptions = getContextOptions(readable(""));
  $: optionsForType = $contextOptions[chatCtx.type] ?? [];
  $: hasOptions = chatCtx.type in $contextOptions;

  $: attrs = convertContextToAttrs(chatCtx);

  $: if (attrs) setTimeout(() => onUpdate(), 50);

  function updateValue(newValue: string) {
    chatCtx.value = newValue;
    chatCtx.label =
      ContextTypeData[chatCtx.type]?.getLabel(chatCtx, metadata) ?? newValue;
  }
</script>

<span
  class="inline-block gap-1 text-sm underline"
  contenteditable="false"
  {...attrs}
>
  <div class="flex flex-row items-center">
    <span>{chatCtx.label}</span>
    {#if hasOptions}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <button
            {...getAttrs([builder])}
            use:builderActions={{ builders: [builder] }}
          >
            <ChevronDownIcon size={12} />
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content>
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
