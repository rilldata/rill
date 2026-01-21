<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  export let disabled = false;
  export let href: string | undefined = undefined;
</script>

{#key disabled}
  <DropdownMenu.Item
    {href}
    class="gap-2"
    aria-disabled={disabled}
    {disabled}
    on:click
  >
    {#if $$slots.icon}
      <div
        class="grid place-content-center opacity-80 text-fg-secondary"
        style:height="18px"
      >
        <slot name="icon" />
      </div>
    {/if}
    <div class:text-fg-disabled={disabled} class="text-left">
      <slot />

      <div
        class:text-fg-secondary={!disabled}
        class:text-fg-disabled={disabled}
        style:font-size="11px"
      >
        <slot name="description" />
      </div>
    </div>

    {#if $$slots["right"]}
      <div
        class="grid place-content-center text-right text-fg-muted"
        style:height="18px"
      >
        <slot name="right" />
      </div>
    {/if}
  </DropdownMenu.Item>
{/key}
