<script lang="ts">
  import AddDropdown from "@rilldata/web-common/features/chat/core/context/AddDropdown.svelte";
  import ChatContext from "@rilldata/web-common/features/chat/core/context/ChatContext.svelte";
  import type { ChatContextEntry } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
  import {
    convertContextToHtml,
    convertContextToInlinePrompt,
    convertHTMLElementToContext,
  } from "@rilldata/web-common/features/chat/core/context/conversions.ts";
  import { getContextMetadataStore } from "@rilldata/web-common/features/chat/core/context/get-context-metadata-store.ts";

  let value: string = "";
  export let onChange: (newValue: string) => void;

  const SPACE_TEXT = "\u00A0";

  let editorElement: HTMLDivElement;
  let isContextMode = false;
  let addContextComponent;
  let addContextNode;

  const contextMetadataStore = getContextMetadataStore();
  $: metadata = $contextMetadataStore;

  export function setPrompt(prompt: string) {
    editorElement.innerHTML = convertContextToHtml(prompt, {
      validExplores: {},
      measures: {},
      dimensions: {},
    });
    setTimeout(() => {
      editorElement.focus();
      findInlineContextNodes(editorElement).forEach(
        ([parent, node, chatCtx]) => {
          new ChatContext({
            target: parent as any,
            anchor: node as any,
            props: {
              chatCtx,
              metadata,
              onUpdate: updateValue,
            },
          });
          parent.removeChild(node);
        },
      );
    });
  }

  export function focusEditor() {
    editorElement.focus();
  }

  function isChildOfEditor(node: Node | null | undefined) {
    if (!node) return false;
    if (editorElement.contains(node)) return true;
    return isChildOfEditor(node.parentNode);
  }

  function insertChatContext(
    componentCreator: (
      target: Element | Document | ShadowRoot,
      anchor: Element | undefined,
    ) => Node | undefined,
  ) {
    const selection = window.getSelection();

    if (
      selection &&
      selection.rangeCount > 0 &&
      isChildOfEditor(selection.anchorNode?.parentNode)
    ) {
      const range = selection.getRangeAt(0);
      range.deleteContents();

      // Use the space node to insert the component
      let node = componentCreator(
        selection.anchorNode?.parentNode as any,
        selection.anchorNode as any,
      );
      node ??= document.createTextNode(SPACE_TEXT);
      range.insertNode(node);

      // Move cursor after the node
      range.setStartAfter(node);
      range.setEndAfter(node);
      selection.removeAllRanges();
      selection.addRange(range);
    } else {
      let node = componentCreator(editorElement, undefined);
      node ??= document.createTextNode(SPACE_TEXT);
      editorElement.appendChild(node);
    }

    updateValue();
  }

  function getValue(node: Node, level: number = 0) {
    if (node.nodeType === Node.TEXT_NODE) {
      return node.textContent;
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      const ctx = convertHTMLElementToContext(node as any, metadata);
      if (ctx) return convertContextToInlinePrompt(ctx);
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

  function findInlineContextNodes(
    node: Node,
  ): [Node, Node, ChatContextEntry][] {
    const inlineContextNodes: [Node, Node, ChatContextEntry][] = [];

    for (const childNode of node.childNodes) {
      if (childNode.nodeName === "DIV") {
        inlineContextNodes.push(...findInlineContextNodes(childNode));
        continue;
      }

      const chatCtx = convertHTMLElementToContext(
        childNode as HTMLElement,
        metadata,
      );
      if (!chatCtx) continue;

      inlineContextNodes.push([node, childNode, chatCtx]);
    }

    return inlineContextNodes;
  }

  function handleKeydown(event: KeyboardEvent) {
    if (isContextMode) {
      setTimeout(() => {
        if (addContextNode && addContextComponent) {
          const searchText = getValue(addContextNode).trim().replace(/^@/, "");
          addContextComponent.setText(searchText);
        }
      });
      return;
    }

    // Detect @ for pill mode (or any other trigger character)
    if (event.key === "@") {
      handleContextStarted();
    }
  }

  function handleContextStarted() {
    isContextMode = true;
    insertChatContext(() => {
      const rect = editorElement.getBoundingClientRect();
      addContextComponent = new AddDropdown({
        target: document.body,
        props: {
          left: rect.left,
          bottom: window.innerHeight - rect.top,
          onAdd: handleContextEnded,
        },
      });
      addContextNode = document.createTextNode(SPACE_TEXT);
      return addContextNode;
    });
  }

  function handleContextEnded(chatCtx: ChatContextEntry) {
    addContextComponent.$destroy();
    isContextMode = false;
    if (!addContextNode?.parentNode) return;
    new ChatContext({
      target: addContextNode.parentNode,
      anchor: addContextNode,
      props: {
        chatCtx,
        metadata,
        onUpdate: updateValue,
      },
    });
    addContextNode.parentNode.removeChild(addContextNode);
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
  on:keydown={handleKeydown}
></div>

<style lang="postcss">
  .chat-input {
    @apply p-2 min-h-[2.5rem] outline-none;
    @apply text-sm leading-relaxed;
  }
</style>
