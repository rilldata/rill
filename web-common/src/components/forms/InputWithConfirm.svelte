<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Pencil from "svelte-radix/Pencil1.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";

  export let value: string | undefined = "";
  export let onConfirm: (newValue: string) => void | Promise<void> = () => {};
  export let id = "";
  export let textClass = "";
  export let editable = true;

  let hovering = false;
  let editing = false;
  let open = false;
  let editedValue = value;

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
  class="h-full w-fit font-medium flex gap-x-2 items-center"
  on:mouseenter={() => (hovering = true)}
  on:mouseleave={() => (hovering = false)}
>
  {#if editing}
    <Input
      {id}
      bind:value={editedValue}
      width="fit"
      height="100%"
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
    <h1
      class="border border-transparent h-full flex justify-center items-center font-medium select-none pointer-events-none pl-2 {textClass}"
    >
      {value}
    </h1>

    {#if editable}
      <DropdownMenu.Root bind:open>
        <DropdownMenu.Trigger asChild let:builder>
          <Button
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
