<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { XIcon } from "lucide-svelte";
  import { slide } from "svelte/transition";

  /**
   * Input that allows to enter multiple items but appears within a single input box.
   */

  export let id: string;
  export let label = "";
  export let description = "";
  export let placeholder = "";
  export let hint = "";
  export let contentClassName = "";

  export let singular: string;
  export let plural: string;
  export let preventFocus = false;

  export let values: string[];
  export let errors: Record<string | number, string[]> | undefined;

  $: lastIdx = values.length - 1;
  $: lastValue = values[lastIdx] ?? "";

  const isMac = window.navigator.userAgent.includes("Macintosh");

  let focused = false;
  function handleKeyDown(event: KeyboardEvent) {
    if (
      // support enter
      (event.key !== "Enter" &&
        // support tab
        event.key !== "Tab" &&
        // support comma
        event.key !== ",") ||
      lastValue === ""
    ) {
      if (event.key === "v" && (isMac ? event.metaKey : event.ctrlKey)) {
        void (async function () {
          // create a scope and wait for input to change when something is pasted
          const prevInput = lastValue;
          await waitUntil(() => prevInput !== lastValue);
          consumeInput();
        })();
      } else if (
        (event.key === "Delete" || event.key === "Backspace") &&
        lastValue === "" &&
        values.length > 1
      ) {
        // remove the last pill when delete/backspace was pressed with empty input
        handleRemove(lastIdx - 1);
      }
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    consumeInput();
  }

  function consumeInput() {
    values = values.slice(0, lastIdx).concat(
      ...lastValue
        .split(",")
        .map((v) => v.trim())
        .filter(Boolean),
      "",
    );
  }

  function handleRemove(index: number) {
    values = values.filter((_, i) => i !== index);
  }

  let error: string;
  $: {
    const errorCount = values
      // ignore the last value which would be being actively edited
      .slice(0, lastIdx)
      .filter((s, i) => s.trim().length > 0 && !!errors?.[i]?.length).length;
    if (errorCount === 0) {
      error = "";
    } else {
      const errorIndex = values.findIndex((_, i) => !!errors?.[i]?.length);
      const firstValue = values[errorIndex];
      if (errorCount === 1) {
        error = `"${firstValue}" is not a valid ${singular}`;
      } else {
        error = `"${firstValue}" and ${errorCount - 1} other${errorCount > 2 ? "s" : ""} are not valid ${plural}`;
      }
    }
  }
  $: hasSomeValue = values[lastIdx].length > 0 || values.length > 1;
  $: hasSomeErrors = !!error;
</script>

<div class="flex flex-col w-full">
  {#if label}
    <div class="flex items-center gap-x-1">
      <label for={id} class="text-fg-primary text-sm font-medium">{label}</label
      >
      {#if hint}
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-fg-secondary" style="transform:translateY(-.5px)">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            {hint}
          </TooltipContent>
        </Tooltip>
      {/if}
      {#if description}
        <div class="text-sm text-fg-secondary">{description}</div>
      {/if}
    </div>
  {/if}
  <div class="flex flex-row gap-1.5 items-center">
    <div
      class="flex flex-row items-center bg-input rounded-sm px-1 py-[3px] w-full {contentClassName}"
      class:border={!hasSomeErrors}
      class:border-gray-300={!hasSomeErrors}
      class:outline={focused || hasSomeErrors}
      class:outline-red-500={hasSomeErrors}
      class:outline-primary-500={focused && !hasSomeErrors}
    >
      <div class="flex flex-wrap gap-1 w-full min-h-[24px]">
        {#each values.slice(0, lastIdx) as _, i (i)}
          {@const hasError = errors?.[i]?.length}
          <div
            class="flex items-center text-fg-secondary text-sm rounded-2xl border bg-input pl-2 pr-1 max-w-full"
            class:border-gray-300={!hasError}
            class:border-red-300={hasError}
            class:bg-red-50={hasError}
          >
            <div
              class="w-fit h-5 overflow-hidden text-ellipsis"
              class:text-red-600={hasError}
            >
              {values[i]}
            </div>
            <IconButton disableHover size={20} on:click={() => handleRemove(i)}>
              <XIcon
                size="12px"
                class="{hasError
                  ? 'text-red-600'
                  : 'text-fg-secondary'} cursor-pointer"
              />
            </IconButton>
          </div>
        {/each}
        <input
          bind:value={values[lastIdx]}
          on:keydown={handleKeyDown}
          autocomplete="off"
          id="{id}.{lastIdx}"
          class="focus:outline-none group-hover:text-red-500 placeholder-fg-secondary text-sm grow px-1 bg-transparent"
          on:focusin={() => (focused = true)}
          on:focusout={() => (focused = false)}
          on:click|preventDefault|stopPropagation={() => preventFocus}
          placeholder={!hasSomeValue ? placeholder : ""}
        />
      </div>
      {#if hasSomeValue}
        <slot name="within-input" />
      {/if}
    </div>
    <slot name="beside-input" {hasSomeValue} />
  </div>
  {#if hasSomeErrors}
    <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px mt-0.5">
      {error}
    </div>
  {/if}
</div>
