<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getContextOptions } from "@rilldata/web-common/features/chat/core/context/context-options.ts";
  import { type ChatContextEntry } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import { getAttrs, builderActions } from "bits-ui";
  import { ChevronDownIcon } from "lucide-svelte";
  import { readable } from "svelte/store";

  export let chatCtx: ChatContextEntry;
  export let onUpdate: () => void;

  const contextOptions = getContextOptions(readable(""));
  $: optionsForType = $contextOptions[chatCtx.type] ?? [];

  $: values = [chatCtx.type, chatCtx.value].concat(
    chatCtx.childEntries?.map((e) => e.value) ?? [],
  );
  $: value = values.join(":");

  $: if (value !== undefined) setTimeout(() => onUpdate(), 50);

  function updateValue(newValue: string) {
    chatCtx.value = newValue;
  }
</script>

<span
  class="inline-block gap-1 text-sm underline"
  contenteditable="false"
  data-value={value}
>
  <div class="flex flex-row items-center">
    <span>{chatCtx.label}</span>
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
  </div>
</span>
