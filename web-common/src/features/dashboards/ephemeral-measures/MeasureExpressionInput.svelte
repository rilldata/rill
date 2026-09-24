<!-- @component
The expression input of the ephemeral measure editors.

Typing "@" opens a picker of the measures the expression may reference,
searchable by display name as well as name (see measure-mention.ts), so users
can insert a measure without knowing its field name.
-->
<script lang="ts">
  import * as Kbd from "@rilldata/web-common/components/kbd";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import {
    autoUpdate,
    computePosition,
    flip,
    offset,
    shift,
  } from "@floating-ui/dom";
  import { ArrowDown, ArrowUp } from "lucide-svelte";
  import { tick } from "svelte";
  import {
    applyMeasureMention,
    filterMeasureMentions,
    findMeasureMention,
  } from "./measure-mention";

  export let value: string;
  export let id: string;
  // The measures the expression may reference.
  export let measures: MetricsViewSpecMeasure[];
  export let errors: string | undefined = undefined;

  let wrapper: HTMLDivElement;
  let cursor = 0;
  let focused = false;
  // Offset of the "@" whose picker was dismissed with Escape, so it stays
  // closed until the user moves to another mention.
  let dismissedStart: number | null = null;
  let focusedIndex = 0;

  $: mention = findMeasureMention(value, cursor);
  $: options = mention ? filterMeasureMentions(measures, mention.query) : [];
  // Typing past every match (e.g. "@total_cost + 5") closes the picker too.
  $: open =
    focused &&
    !!mention &&
    mention.start !== dismissedStart &&
    options.length > 0;
  $: if (options) focusedIndex = 0;
  // The "@" is not valid in an expression, so the parse error it causes is
  // noise while the picker is open.
  $: shownErrors = open ? undefined : errors;

  function inputElement(): HTMLInputElement | null {
    return wrapper?.querySelector("input") ?? null;
  }

  function syncCursor() {
    cursor = inputElement()?.selectionStart ?? value.length;
  }

  function onKeydown(e: KeyboardEvent) {
    if (!open) return;
    switch (e.key) {
      case "ArrowDown":
        focusedIndex = (focusedIndex + 1) % options.length;
        break;
      case "ArrowUp":
        focusedIndex = (focusedIndex - 1 + options.length) % options.length;
        break;
      case "Enter":
        void select(options[focusedIndex]);
        break;
      case "Escape":
        dismissedStart = mention?.start ?? null;
        break;
      default:
        return;
    }
    // Keep the key from the input and from the dialog, which closes on Escape.
    e.preventDefault();
    e.stopPropagation();
  }

  async function select(measure: MetricsViewSpecMeasure) {
    if (!mention || !measure.name) return;
    const applied = applyMeasureMention(value, mention, measure.name);
    value = applied.value;
    cursor = applied.cursor;
    await tick();
    const input = inputElement();
    if (!input) return;
    input.focus();
    input.setSelectionRange(applied.cursor, applied.cursor);
  }

  function position(list: HTMLElement) {
    const input = inputElement();
    if (!input) return;
    const cleanup = autoUpdate(input, list, () => {
      list.style.width = `${input.getBoundingClientRect().width}px`;
      void computePosition(input, list, {
        strategy: "fixed",
        placement: "bottom-start",
        middleware: [offset(4), flip(), shift({ padding: 8 })],
      }).then(({ x, y }) => {
        Object.assign(list.style, { left: `${x}px`, top: `${y}px` });
      });
    });
    return { destroy: cleanup };
  }
</script>

<!-- Keys are intercepted in the capture phase so the picker handles them
     before the input's own handlers (which blur on Enter). Clicks and key
     presses within the input move the cursor, which decides the mention. -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
  bind:this={wrapper}
  class="relative"
  on:keydown|capture={onKeydown}
  on:keyup={syncCursor}
  on:click={syncCursor}
  on:focusin={() => (focused = true)}
  on:focusout={() => (focused = false)}
>
  <Input
    bind:value
    {id}
    label={m.dashboard_pivot_ephemeral_expression_label()}
    placeholder={m.dashboard_pivot_ephemeral_expression_placeholder()}
    hint={m.dashboard_pivot_ephemeral_expression_hint()}
    errors={shownErrors}
    alwaysShowError
    oninput={syncCursor}
  />

  {#if open}
    <div
      use:position
      role="listbox"
      aria-label={m.dashboard_pivot_ephemeral_mention_label()}
      class="measure-picker"
    >
      <div class="options">
        {#each options as measure, i (measure.name)}
          <button
            type="button"
            role="option"
            aria-selected={i === focusedIndex}
            class="option"
            class:focused={i === focusedIndex}
            on:mousemove={() => (focusedIndex = i)}
            on:mousedown|preventDefault
            on:click={() => select(measure)}
          >
            <span class="truncate">{measure.displayName || measure.name}</span>
            {#if measure.displayName && measure.displayName !== measure.name}
              <span class="ml-auto pl-2 text-fg-muted truncate">
                {measure.name}
              </span>
            {/if}
          </button>
        {/each}
      </div>
      <div class="navigation">
        <Kbd.Group>
          <Kbd.Root><ArrowUp size="12px" /></Kbd.Root>
          <Kbd.Root><ArrowDown size="12px" /></Kbd.Root>
          <span>{m.dashboard_pivot_ephemeral_mention_navigate()}</span>
          <!-- i18n-ignore: key name -->
          <Kbd.Root><span>Enter</span></Kbd.Root>
          <span>{m.dashboard_pivot_ephemeral_mention_select()}</span>
        </Kbd.Group>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  .measure-picker {
    @apply fixed top-0 left-0 z-50 flex flex-col p-1.5;
    @apply border rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .options {
    @apply flex flex-col max-h-48 overflow-y-auto;
  }

  .option {
    @apply flex items-center w-full px-2 py-1.5 rounded-sm text-xs text-left;
  }

  .option.focused {
    @apply bg-popover-accent;
  }

  .navigation {
    @apply flex items-center pt-1.5 px-1.5 text-xs text-popover-foreground/60;
  }
</style>
