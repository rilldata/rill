<script lang="ts">
  import { convertContextToHtml } from "@rilldata/web-common/features/chat/core/context/conversions.ts";
  import { ChatInputTextAreaManager } from "@rilldata/web-common/features/chat/core/input/chat-input-textarea-manager.ts";

  export let onChange: (newValue: string) => void;
  export let onSubmit: () => void;

  const placeholder = "Ask about your data...";
  let editorElement: HTMLDivElement;
  let value = "";

  const manager = new ChatInputTextAreaManager();
  $: manager.setElement(editorElement);
  $: manager.setOnChange((newVal) => {
    value = newVal;
    onChange(newVal);
  });
  $: manager.setOnSubmit(onSubmit);

  export function setPrompt(prompt: string) {
    value = prompt;
    const html = convertContextToHtml(prompt, {
      validExplores: {},
      measures: {},
      dimensions: {},
    });
    manager.setHtml(html);
  }

  export function focusEditor() {
    manager?.focusEditor();
  }
</script>

<div
  bind:this={editorElement}
  contenteditable="true"
  role="textbox"
  tabindex="0"
  class="chat-input"
  data-placeholder={placeholder}
  class:empty={!value.length}
  on:input={manager.updateValue}
  on:keydown={manager.handleKeydown}
></div>

<style lang="postcss">
  .chat-input {
    @apply p-2 min-h-[2.5rem] outline-none;
    @apply text-sm leading-relaxed;
  }

  .chat-input.empty:before {
    content: attr(data-placeholder);
    @apply text-gray-400 pointer-events-none absolute;
  }
</style>
