<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Pencil from "svelte-radix/Pencil1.svelte";
  import { cubicOut } from "svelte/easing";
  import { scale } from "svelte/transition";

  export let value: string | undefined = "";
  export let onConfirm: (newValue: string) => void | Promise<void> = () => {};
  export let id = "";
  export let textClass = "";
  export let editable = true;
  export let bumpDown = false;
  export let type: "Project" | "File" = "File";
  export let size: "sm" | "md" | "lg" = "lg";
  export let showIndicator = false;
  export let editing = false;

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
  class="h-full w-fit font-medium flex gap-x-1.5 items-center group"
  class:w-full={editing}
  class:truncate={!editing}
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
      onClick={triggerConfirm}
    >
      <Check size="16px" />
    </Button>
  {:else}
    <div class="input-wrapper truncate">
      <h1 class:bump-down={bumpDown} class={textClass} title={value}>
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
      <span class="group-hover:flex hidden">
        <Button
          label="{type} title actions"
          square
          small
          type="ghost"
          onClick={() => {
            editing = !editing;
          }}
        >
          <Pencil size="16px" />
        </Button>
      </span>
    {/if}
  {/if}
</div>

<style lang="postcss">
  h1 {
    @apply p-0 bg-transparent;
    @apply w-full max-w-fit;
    @apply outline-none border-0;
    @apply cursor-default truncate;
  }

  .bump-down {
    @apply pt-[1px];
  }

  .input-wrapper {
    @apply overflow-hidden;
    @apply flex justify-center items-center pl-2;
    @apply justify-center;
    @apply border border-transparent rounded-[2px];
    @apply h-fit;
  }
</style>
