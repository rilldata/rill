<script lang="ts">
  import { convertContextToHtml } from "@rilldata/web-common/features/chat/core/context/conversions.ts";
  import { ChatInputTextAreaManager } from "@rilldata/web-common/features/chat/core/input/chat-input-textarea-manager.ts";

  export let onChange: (newValue: string) => void;

  let editorElement: HTMLDivElement;

  const manager = new ChatInputTextAreaManager();
  $: manager.setElement(editorElement);
  $: manager.setOnChange(onChange);

  export function setPrompt(prompt: string) {
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
  class:empty={!editorElement?.textContent?.trim()}
  on:input={manager.updateValue}
  on:keydown={manager.handleKeydown}
></div>

<style lang="postcss">
  .chat-input {
    @apply p-2 min-h-[2.5rem] outline-none;
    @apply text-sm leading-relaxed;
  }
</style>
