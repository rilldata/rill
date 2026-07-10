<script lang="ts">
  import Close from "@rilldata/web-common/components/icons/Close.svelte";

  type Props = {
    tags: string[];
    suggestions?: string[];
    label?: string;
    placeholder?: string;
    onChange: (tags: string[]) => void;
  };

  let {
    tags,
    suggestions = [],
    label = "Tags",
    placeholder = "Add a tag and press Enter",
    onChange,
  }: Props = $props();

  let inputValue = $state("");
  let inputRef = $state<HTMLInputElement | undefined>(undefined);
  let wrapperRef = $state<HTMLDivElement | undefined>(undefined);
  let dropdownEl = $state<HTMLDivElement | undefined>(undefined);
  let dropdownOpen = $state(false);
  let highlightedIndex = $state(0);
  let dropdownPos = $state<{
    top?: number;
    bottom?: number;
    left: number;
    width: number;
    maxHeight: number;
    placement: "below" | "above";
  }>({
    left: 0,
    width: 0,
    maxHeight: 240,
    placement: "below",
  });

  const DROPDOWN_OFFSET = 4;
  const DROPDOWN_MAX_HEIGHT = 240;
  const VIEWPORT_PADDING = 8;

  let trimmedInput = $derived(inputValue.trim());

  let filteredSuggestions = $derived.by(() => {
    const lowered = trimmedInput.toLowerCase();
    return suggestions.filter(
      (s) =>
        s &&
        !tags.includes(s) &&
        (lowered === "" || s.toLowerCase().includes(lowered)),
    );
  });

  let showDropdown = $derived(dropdownOpen && filteredSuggestions.length > 0);

  $effect(() => {
    if (highlightedIndex >= filteredSuggestions.length) {
      highlightedIndex = 0;
    }
  });

  function updatePosition() {
    if (!wrapperRef) return;
    const rect = wrapperRef.getBoundingClientRect();
    const spaceBelow =
      window.innerHeight - rect.bottom - DROPDOWN_OFFSET - VIEWPORT_PADDING;
    const spaceAbove = rect.top - DROPDOWN_OFFSET - VIEWPORT_PADDING;
    const placement: "below" | "above" =
      spaceBelow >= DROPDOWN_MAX_HEIGHT || spaceBelow >= spaceAbove
        ? "below"
        : "above";
    const maxHeight = Math.max(
      80,
      Math.min(
        DROPDOWN_MAX_HEIGHT,
        placement === "below" ? spaceBelow : spaceAbove,
      ),
    );
    if (placement === "below") {
      dropdownPos = {
        top: rect.bottom + DROPDOWN_OFFSET,
        left: rect.left,
        width: rect.width,
        maxHeight,
        placement,
      };
    } else {
      // Anchor from the bottom so the dropdown stays glued to the input as
      // the filtered list shrinks.
      dropdownPos = {
        bottom: window.innerHeight - rect.top + DROPDOWN_OFFSET,
        left: rect.left,
        width: rect.width,
        maxHeight,
        placement,
      };
    }
  }

  // Recompute when the list of suggestions changes (e.g. typing filters it).
  $effect(() => {
    void filteredSuggestions.length;
    if (showDropdown) updatePosition();
  });

  $effect(() => {
    if (!showDropdown) return;
    updatePosition();
    const handler = () => updatePosition();
    window.addEventListener("scroll", handler, true);
    window.addEventListener("resize", handler);
    return () => {
      window.removeEventListener("scroll", handler, true);
      window.removeEventListener("resize", handler);
    };
  });

  function portal(node: HTMLElement) {
    document.body.appendChild(node);
    return {
      destroy() {
        if (node.parentNode === document.body) {
          document.body.removeChild(node);
        }
      },
    };
  }

  function addTag(value: string) {
    const trimmed = value.trim();
    if (!trimmed || tags.includes(trimmed)) return;
    onChange([...tags, trimmed]);
  }

  function commitInput() {
    const trimmed = trimmedInput.replace(/,$/, "").trim();
    inputValue = "";
    if (trimmed) addTag(trimmed);
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "ArrowDown" && showDropdown) {
      e.preventDefault();
      highlightedIndex = (highlightedIndex + 1) % filteredSuggestions.length;
      return;
    }
    if (e.key === "ArrowUp" && showDropdown) {
      e.preventDefault();
      highlightedIndex =
        (highlightedIndex - 1 + filteredSuggestions.length) %
        filteredSuggestions.length;
      return;
    }
    if (e.key === "Enter") {
      e.preventDefault();
      if (showDropdown && filteredSuggestions[highlightedIndex]) {
        addTag(filteredSuggestions[highlightedIndex]);
        inputValue = "";
      } else {
        commitInput();
      }
      return;
    }
    if (e.key === ",") {
      e.preventDefault();
      commitInput();
      return;
    }
    if (e.key === "Escape" && dropdownOpen) {
      e.preventDefault();
      dropdownOpen = false;
      return;
    }
    if (e.key === "Backspace" && inputValue === "" && tags.length > 0) {
      e.preventDefault();
      onChange(tags.slice(0, -1));
    }
  }

  function pickSuggestion(s: string) {
    addTag(s);
    inputValue = "";
    inputRef?.focus();
  }

  function removeTag(t: string) {
    onChange(tags.filter((x) => x !== t));
    inputRef?.focus();
  }

  function handleFocusIn() {
    dropdownOpen = true;
    highlightedIndex = 0;
  }

  function handleInputBlur(e: FocusEvent) {
    // Keep the dropdown open if focus moved into the portaled dropdown
    // (e.g. clicking a suggestion). Close otherwise.
    const next = e.relatedTarget as Node | null;
    if (next && dropdownEl?.contains(next)) return;
    commitInput();
    dropdownOpen = false;
  }
</script>

<div class="flex flex-col gap-y-1" bind:this={wrapperRef}>
  <span class="text-fg-secondary text-sm font-medium">{label}</span>
  <div
    class="input-wrapper flex flex-wrap items-center gap-1 px-1.5 py-1 min-h-8 cursor-text"
    role="presentation"
    onclick={() => inputRef?.focus()}
  >
    {#each tags as t (t)}
      <span
        class="inline-flex items-center gap-x-1 pl-2 pr-1 py-0.5 text-xs rounded-sm bg-primary-50 text-fg-primary border border-primary-100"
      >
        {t}
        <button
          type="button"
          class="text-fg-secondary hover:text-fg-primary p-0.5 rounded-sm"
          onclick={(e) => {
            e.stopPropagation();
            removeTag(t);
          }}
          aria-label={`Remove ${t}`}
        >
          <Close size="10px" />
        </button>
      </span>
    {/each}
    <input
      bind:this={inputRef}
      bind:value={inputValue}
      onkeydown={handleKeyDown}
      onfocus={handleFocusIn}
      onblur={handleInputBlur}
      class="flex-1 min-w-[100px] bg-transparent outline-none text-sm py-0.5"
      {placeholder}
      autocomplete="off"
      aria-label={label}
      aria-autocomplete="list"
      aria-expanded={showDropdown}
    />
  </div>
</div>

{#if showDropdown}
  <div
    use:portal
    bind:this={dropdownEl}
    class="fixed z-popover rounded-md border bg-popover text-popover-foreground shadow-md overflow-y-auto py-1"
    style:top={dropdownPos.top !== undefined ? `${dropdownPos.top}px` : "auto"}
    style:bottom={dropdownPos.bottom !== undefined
      ? `${dropdownPos.bottom}px`
      : "auto"}
    style:left="{dropdownPos.left}px"
    style:width="{dropdownPos.width}px"
    style:max-height="{dropdownPos.maxHeight}px"
    role="listbox"
  >
    {#each filteredSuggestions as s, i (s)}
      <button
        type="button"
        role="option"
        aria-selected={highlightedIndex === i}
        class="w-full text-left px-2.5 py-1 text-sm text-fg-primary hover:bg-popover-accent flex items-center"
        class:bg-popover-accent={highlightedIndex === i}
        onmousedown={(e) => e.preventDefault()}
        onclick={() => pickSuggestion(s)}
        onmouseenter={() => (highlightedIndex = i)}
      >
        {#if trimmedInput && s
            .toLowerCase()
            .includes(trimmedInput.toLowerCase())}
          {@const idx = s.toLowerCase().indexOf(trimmedInput.toLowerCase())}
          <span>{s.slice(0, idx)}</span>
          <span class="font-semibold text-fg-primary"
            >{s.slice(idx, idx + trimmedInput.length)}</span
          >
          <span>{s.slice(idx + trimmedInput.length)}</span>
        {:else}
          {s}
        {/if}
      </button>
    {/each}
  </div>
{/if}

<style lang="postcss">
  .input-wrapper {
    @apply border border-gray-300 rounded-[2px] bg-input;
  }
  .input-wrapper:focus-within {
    @apply border-primary-500 ring-2 ring-primary-100;
  }
</style>
