<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Pencil from "svelte-radix/Pencil1.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { scale } from "svelte/transition";
  import { cubicOut } from "svelte/easing";

  export let value: string | undefined = "";
  export let onConfirm: (newValue: string) => void | Promise<void> = () => {};
  export let id = "";
  export let textClass = "";
  export let editable = true;
  export let bumpDown = false;
  export let type: "Project" | "File" = "File";
  export let size: "sm" | "md" | "lg" = "lg";
  export let showIndicator = false;

  let hovering = false;
  let editing = false;
  let open = false;

  $: editedValue = value;

  async function triggerConfirm() {
    if (!editedValue) return;
    await onConfirm(editedValue);
    reset();
  }

  function reset() {
    editedValue = value;
    editing = false;
  }
</script>

<div
  role="presentation"
  class="h-full w-fit font-medium flex gap-x-0 items-center"
  on:mouseenter={() => (hovering = true)}
  on:mouseleave={() => (hovering = false)}
>
  {#if editing}
    <Input
      {size}
      {id}
      bind:value={editedValue}
      claimFocusOnMount
      onEnter={triggerConfirm}
      onEscape={reset}
      {textClass}
      onBlur={(e) => {
        const target = e.relatedTarget;
        if (
          target instanceof HTMLElement &&
          target.getAttribute("aria-label") === "Save title"
        ) {
          return;
        }

        reset();
      }}
    />

    <Button
      type="ghost"
      small
      square
      label="Save title"
      on:click={triggerConfirm}
    >
      <Check size="16px" />
    </Button>
  {:else}
    <div class="input-wrapper">
      <h1 class:bump-down={bumpDown} class="{textClass} {size}">
        {value}
      </h1>
    </div>

    {#if showIndicator}
      <span
        class="w-1.5 h-1.5 bg-gray-300 rounded flex-none mr-1"
        transition:scale={{ duration: 200, easing: cubicOut }}
      />
    {/if}

    {#if editable}
      <DropdownMenu.Root bind:open>
        <DropdownMenu.Trigger asChild let:builder>
          <Button
            label="{type} title actions"
            builders={[builder]}
            square
            small
            type="ghost"
            class={hovering ? "" : "opacity-0 pointer-events-none"}
          >
            <ThreeDot size="16px" />
          </Button>
        </DropdownMenu.Trigger>

        <DropdownMenu.Content align="start">
          <DropdownMenu.Item
            on:click={() => {
              editing = !editing;
            }}
          >
            <Pencil size="16px" />
            Rename
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  {/if}
</div>

<style lang="postcss">
  h1 {
    @apply flex items-center;
    @apply p-0 bg-transparent;
    @apply size-full;
    @apply outline-none border-0;
    @apply cursor-default min-w-fit;
    vertical-align: middle;
  }

  .bump-down {
    @apply pt-[1px];
  }

  .input-wrapper {
    @apply overflow-hidden;
    @apply flex justify-center items-center px-2;
    @apply w-fit  justify-center;
    @apply border border-transparent rounded-[2px];
    @apply h-fit;
  }

  .sm {
    height: 24px;
  }

  .md {
    height: 26px;
  }

  .lg {
    height: 30px;
  }
</style>
