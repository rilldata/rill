<script context="module" lang="ts">
  export type FormInputEvent<T extends Event = Event> = T & {
    currentTarget: EventTarget & HTMLInputElement;
  };
  export type InputEvents = {
    blur: FormInputEvent<FocusEvent>;
    change: FormInputEvent<Event>;
    click: FormInputEvent<MouseEvent>;
    focus: FormInputEvent<FocusEvent>;
    focusin: FormInputEvent<FocusEvent>;
    focusout: FormInputEvent<FocusEvent>;
    keydown: FormInputEvent<KeyboardEvent>;
    keypress: FormInputEvent<KeyboardEvent>;
    keyup: FormInputEvent<KeyboardEvent>;
    mouseover: FormInputEvent<MouseEvent>;
    mouseenter: FormInputEvent<MouseEvent>;
    mouseleave: FormInputEvent<MouseEvent>;
    mousemove: FormInputEvent<MouseEvent>;
    paste: FormInputEvent<ClipboardEvent>;
    input: FormInputEvent<InputEvent>;
    wheel: FormInputEvent<WheelEvent>;
  };
</script>

<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";
  import { cn } from "@rilldata/web-common/lib/shadcn.ts";

  type $$Props = HTMLInputAttributes & { files?: FileList };

  let className: $$Props["class"] = undefined;
  export let value: $$Props["value"] = undefined;
  export let files: $$Props["files"] = undefined;
  export let type: $$Props["type"];
  export { className as class };

  // Workaround for https://github.com/sveltejs/svelte/issues/9305
  // Fixed in Svelte 5, but not backported to 4.x.
  export let readonly: $$Props["readonly"] = undefined;

  const classes = cn(
    "border-input placeholder:text-muted-foreground focus-visible:ring-ring flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:outline-none focus-visible:ring-1 disabled:cursor-not-allowed disabled:opacity-50",
    className,
  );
</script>

{#if type === "file"}
  <input
    type="file"
    class={classes}
    bind:value
    bind:files
    {readonly}
    on:blur
    on:change
    on:click
    on:focus
    on:focusin
    on:focusout
    on:keydown
    on:keypress
    on:keyup
    on:mouseover
    on:mouseenter
    on:mouseleave
    on:mousemove
    on:paste
    on:input
    on:wheel|passive
    {...$$restProps}
  />
{:else}
  <input
    class={classes}
    bind:value
    {readonly}
    on:blur
    on:change
    on:click
    on:focus
    on:focusin
    on:focusout
    on:keydown
    on:keypress
    on:keyup
    on:mouseover
    on:mouseenter
    on:mouseleave
    on:mousemove
    on:paste
    on:input
    on:wheel|passive
    {...$$restProps}
  />
{/if}
