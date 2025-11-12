<script lang="ts">
  import type { ChatContext } from "@rilldata/web-common/features/chat/core/input/types.ts";
  import ChatContextComponent from "@rilldata/web-common/features/chat/core/input/ChatContext.svelte";

  export let value: string = "";
  export let onChange: (newValue: string) => void;

  let editorElement: HTMLDivElement;

  function isChildOfEditor(node: Node | undefined) {
    if (!node) return false;
    if (editorElement.contains(node)) return true;
    return isChildOfEditor(node.parentNode);
  }

  export function insertChatContext(chatCtx: ChatContext) {
    const selection = window.getSelection();

    if (
      selection &&
      selection.rangeCount > 0 &&
      isChildOfEditor(selection.anchorNode?.parentNode)
    ) {
      const range = selection.getRangeAt(0);
      range.deleteContents();

      // Add a space after the pill
      const space = document.createTextNode("\u00A0");
      range.insertNode(space);

      // Use the space node to insert the component
      new ChatContextComponent({
        target: selection.anchorNode?.parentNode as any,
        anchor: space as any,
        props: {
          chatCtx,
        },
      });

      // Move cursor after the space
      range.setStartAfter(space);
      range.setEndAfter(space);
      selection.removeAllRanges();
      selection.addRange(range);
    } else {
      new ChatContextComponent({
        target: editorElement,
        props: {
          chatCtx,
        },
      });
      editorElement.appendChild(document.createTextNode("\u00A0"));
    }

    updateValue();
  }

  function getValue(node: Node, level: number = 0) {
    if (node.nodeType === Node.TEXT_NODE) {
      return node.textContent;
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      const val = (node as any).getAttribute("data-value");
      if (val) return val;
      const prefix = node.nodeName === "DIV" && level !== 0 ? "\n" : "";
      return (
        prefix +
        Array.from(node.childNodes)
          .map((c) => getValue(c, level + 1))
          .join("")
      );
    }
    return "";
  }

  function updateValue() {
    value = getValue(editorElement);
    onChange(value);
  }
</script>

<div
  bind:this={editorElement}
  contenteditable="true"
  role="textbox"
  tabindex="0"
  class="chat-input"
  class:empty={!editorElement?.textContent?.trim()}
  on:input={updateValue}
></div>

<style lang="postcss">
  .chat-input {
    @apply p-2 min-h-[2.5rem] outline-none;
    @apply text-sm leading-relaxed;
  }
</style>
