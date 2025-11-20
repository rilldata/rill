import AddDropdown from "@rilldata/web-common/features/chat/core/context/AddDropdown.svelte";
import ChatContext from "@rilldata/web-common/features/chat/core/context/ChatContext.svelte";
import {
  type ChatContextEntry,
  ChatContextEntryType,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import {
  convertContextToInlinePrompt,
  convertContextValueToEntry,
  PROMPT_INLINE_CONTEXT_TAG,
} from "@rilldata/web-common/features/chat/core/context/conversions.ts";
import { getContextMetadataStore } from "@rilldata/web-common/features/chat/core/context/get-context-metadata-store.ts";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { get } from "svelte/store";

const GENERAL_CONTEXT_TRIGGER = "@";
const MEASURES_CONTEXT_TRIGGER = "$";
const DIMENSIONS_CONTEXT_TRIGGER = "#";
const DIMENSION_VALUES_CONTEXT_TRIGGER = ":";

export class ChatInputTextAreaManager {
  private editorElement: HTMLDivElement;

  private isContextMode = false;
  private addContextComponent: AddDropdown | null = null;
  private addContextNode: Node | null = null;
  private contextTriggerChar: string = "";
  private elementToContextComponent = new Map<Node, ChatContext>();

  private readonly exploreNameStore = getExploreNameStore();
  private readonly contextMetadataStore = getContextMetadataStore();

  public constructor(
    private readonly onChange: (newValue: string) => void,
    private readonly onSubmit: () => void,
  ) {}

  public setElement(editorElement: HTMLDivElement) {
    this.editorElement = editorElement;
  }

  public setPrompt(html: string) {
    this.editorElement.innerHTML = html;
    setTimeout(() => {
      // Cleanup any old components
      this.elementToContextComponent.values().forEach((c) => c.$destroy());

      const inlineContextNodes = this.findInlineContextNodes(
        this.editorElement,
      );
      inlineContextNodes.forEach(([parent, node, chatCtx]) => {
        const comp = new ChatContext({
          target: parent as any,
          anchor: node as any,
          props: {
            chatCtx,
            metadata: get(this.contextMetadataStore),
            onUpdate: this.updateValue,
          },
        });
        this.componentAdded(comp, node);
        parent.removeChild(node);
      });

      this.editorElement.focus();

      setTimeout(this.updateValue);
    });
  }

  public handleKeydown = (event: KeyboardEvent) => {
    if (!get(this.exploreNameStore)) return; // Only supported within an explore right now.

    setTimeout(() => {
      if (event.key === "Backspace") {
        this.removeNodes();
      } else if (event.key === "Escape" && this.isContextMode) {
        this.removeAddContextComponent(true);
      }

      if (this.isContextMode) {
        if (event.key === "Tab") {
          this.addContextComponent?.selectFirst();
        } else {
          this.handleContextMode();
        }
        return;
      }

      // Detect @ for pill mode (or any other trigger character)
      if (event.key === GENERAL_CONTEXT_TRIGGER) {
        this.handleContextStarted(GENERAL_CONTEXT_TRIGGER);
      } else if (event.key === DIMENSION_VALUES_CONTEXT_TRIGGER) {
        this.handleContextStarted(DIMENSION_VALUES_CONTEXT_TRIGGER);
      } else if (event.key === MEASURES_CONTEXT_TRIGGER) {
        this.handleContextStarted(MEASURES_CONTEXT_TRIGGER);
      } else if (event.key === DIMENSIONS_CONTEXT_TRIGGER) {
        this.handleContextStarted(DIMENSIONS_CONTEXT_TRIGGER);
      } else if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault();
        this.onSubmit();
      }
    });
  };

  public updateValue = () => {
    const value = this.getValue(this.editorElement);
    this.onChange(value);
  };

  public focusEditor = () => {
    this.editorElement.focus();
  };

  private getValue(node: Node, level: number = 0): string {
    if (node.nodeType === Node.TEXT_NODE) {
      return node.textContent ?? "";
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      const comp = this.elementToContextComponent.get(node);
      if (comp) {
        const ctx = comp.getChatContext();
        return convertContextToInlinePrompt(ctx);
      }

      const prefix = node.nodeName === "DIV" && level !== 0 ? "\n" : "";
      return (
        prefix +
        Array.from(node.childNodes)
          .map((c) => this.getValue(c, level + 1))
          .join("")
      );
    }

    return "";
  }

  private findInlineContextNodes(node: Node): [Node, Node, ChatContextEntry][] {
    const inlineContextNodes: [Node, Node, ChatContextEntry][] = [];

    for (const childNode of node.childNodes) {
      if (childNode.nodeName === "DIV") {
        inlineContextNodes.push(...this.findInlineContextNodes(childNode));
        continue;
      }

      if (childNode.nodeName.toLowerCase() !== PROMPT_INLINE_CONTEXT_TAG) {
        continue;
      }

      const chatCtx = convertContextValueToEntry(
        (childNode as HTMLElement).innerText,
        get(this.contextMetadataStore),
      );
      if (!chatCtx) continue;

      inlineContextNodes.push([node, childNode, chatCtx]);
    }

    return inlineContextNodes;
  }

  private handleContextMode() {
    if (this.addContextNode && this.addContextComponent) {
      const contextNodeValue: string = this.getValue(
        this.addContextNode,
      ).trim();
      if (contextNodeValue.length === 0) {
        this.removeAddContextComponent();
      } else {
        const searchText = contextNodeValue.replace(
          new RegExp(`.*?${maybeEscapeTriggerChar(this.contextTriggerChar)}`),
          "",
        );
        this.addContextComponent.setText(searchText);
      }
    }
  }

  private handleContextStarted(contextTriggerChar: string) {
    const selection = window.getSelection();
    if (!selection) return;
    const range = selection.getRangeAt(0);

    const node = range.startContainer;
    if (!node) return;

    let chatCtx: ChatContextEntry | null = null;
    let comp: ChatContext | null = null;

    switch (contextTriggerChar) {
      case MEASURES_CONTEXT_TRIGGER:
        chatCtx = { type: ChatContextEntryType.Measures } as ChatContextEntry;
        break;

      case DIMENSIONS_CONTEXT_TRIGGER:
        chatCtx = { type: ChatContextEntryType.Dimensions } as ChatContextEntry;
        break;

      case DIMENSION_VALUES_CONTEXT_TRIGGER:
        {
          const notDirectlyBesideComponent =
            range.startOffset !== 1 || !(node as any)?.previousElementSibling;
          if (notDirectlyBesideComponent) return;

          const prevSibling = (range.startContainer as any)
            ?.previousElementSibling as HTMLElement;
          comp = this.elementToContextComponent.get(prevSibling) ?? null;
          if (!comp) return;

          chatCtx = comp.getChatContext();
          if (chatCtx?.type !== ChatContextEntryType.Dimensions) return;
          chatCtx.type = ChatContextEntryType.DimensionValue;
          console.log(chatCtx);
        }
        break;
    }

    const rect = this.editorElement.getBoundingClientRect();
    this.addContextComponent = new AddDropdown({
      target: document.body,
      props: {
        left: rect.left,
        bottom: window.innerHeight - rect.top,
        chatCtx,
        onAdd: (ctx) => {
          comp?.$destroy();
          this.handleContextEnded(ctx);
        },
        focusEditor: this.focusEditor,
      },
    });
    this.isContextMode = true;
    this.addContextNode = node;
    this.contextTriggerChar = contextTriggerChar;
  }

  private handleContextEnded = (chatCtx: ChatContextEntry) => {
    if (!this.addContextNode?.parentNode) return;

    const comp = new ChatContext({
      target: this.addContextNode.parentNode as any,
      anchor: this.addContextNode as any,
      props: {
        chatCtx,
        metadata: get(this.contextMetadataStore),
        onUpdate: this.updateValue,
      },
    });

    // Wait a loop to ensure the component is added to the DOM.
    setTimeout(() => {
      this.componentAdded(comp, this.addContextNode!);

      const selection = window.getSelection();
      const range = selection?.getRangeAt(0);
      if (selection && range) {
        range.setStartBefore(this.addContextNode!);
        range.setEndBefore(this.addContextNode!);
        selection.removeAllRanges();
        selection.addRange(range);
      }

      this.removeAddContextComponent();
    });
  };

  private removeAddContextComponent(keepTextNode = false) {
    this.addContextComponent?.$destroy();
    this.isContextMode = false;
    if (keepTextNode || !this.addContextNode) return;

    // Remove the search text after context char.
    this.addContextNode.textContent =
      this.addContextNode.textContent?.replace(
        new RegExp(`${maybeEscapeTriggerChar(this.contextTriggerChar)}.*$`),
        "",
      ) ?? "";
    // Only remove the text node if it's empty after removing the search text.
    // If the search was started in the middle of a line, there will be other text.
    if (this.addContextNode.textContent.length === 0)
      (this.addContextNode as Element).remove();
  }

  private removeNodes() {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return;

    const range = selection.getRangeAt(0);
    let node: Node | null = range.startContainer;
    if (!node || selection.isCollapsed) return;

    do {
      if (node === this.addContextNode) {
        this.addContextComponent?.$destroy();
        this.isContextMode = false;
        this.addContextNode = null;
        this.addContextComponent = null;
      } else {
        const comp = this.elementToContextComponent.get(node);
        comp?.$destroy();
        this.elementToContextComponent.delete(node);
      }

      node = node.nextSibling;
    } while (node && node !== range.endContainer.nextSibling);
  }

  private componentAdded(comp: ChatContext, nextNode: Node) {
    const node = (nextNode as any).previousElementSibling;
    this.elementToContextComponent.set(node, comp);
    // Remove the comment for HMR, it interferes with text editing. This is only added in dev mode.
    if (nextNode.previousSibling?.nodeType === Node.COMMENT_NODE) {
      nextNode.previousSibling.remove();
    }
  }
}

function maybeEscapeTriggerChar(triggerChar: string) {
  const prefix = triggerChar === MEASURES_CONTEXT_TRIGGER ? "\\" : "";
  return `${prefix}${triggerChar}`;
}
