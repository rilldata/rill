import AddDropdown from "@rilldata/web-common/features/chat/core/context/AddDropdown.svelte";
import AddValueDropdown from "@rilldata/web-common/features/chat/core/context/AddValueDropdown.svelte";
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

export class ChatInputTextAreaManager {
  private editorElement: HTMLDivElement;

  private isContextMode = false;
  private addContextComponent;
  private addContextNode;
  private addContextChar: string = "";
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
        this.handleContextMode();
        return;
      }

      // Detect @ for pill mode (or any other trigger character)
      if (event.key === "@") {
        this.handleContextStarted();
      } else if (event.key === ":") {
        this.handleContextSecondValue();
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
          new RegExp(`.*?${this.addContextChar}`),
          "",
        );
        this.addContextComponent.setText(searchText);
      }
    }
  }

  private handleContextSecondValue() {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return;

    const range = selection.getRangeAt(0);
    const notDirectlyBesideComponent =
      range.startOffset !== 1 ||
      !(range.startContainer as any)?.previousElementSibling;
    if (notDirectlyBesideComponent) return;

    const prevSibling = (range.startContainer as any)
      ?.previousElementSibling as HTMLElement;
    const comp = this.elementToContextComponent.get(prevSibling);
    if (!comp) return;

    const ctx = comp.getChatContext();
    if (ctx?.type !== ChatContextEntryType.Dimensions) return;
    ctx.type = ChatContextEntryType.DimensionValue;

    const rect = this.editorElement.getBoundingClientRect();
    this.addContextComponent = new AddValueDropdown({
      target: document.body,
      props: {
        left: rect.left,
        bottom: window.innerHeight - rect.top,
        chatCtx: ctx,
        metadata: get(this.contextMetadataStore),
        onAdd: (ctx) => {
          comp.$destroy();
          this.handleContextEnded(ctx);
        },
      },
    });
    this.addContextNode = range.startContainer;
    this.isContextMode = true;
    this.addContextChar = ":";
  }

  private handleContextStarted() {
    this.isContextMode = true;
    const selection = window.getSelection();
    const anchorNode = selection?.anchorNode;
    if (!anchorNode) return;

    const rect = this.editorElement.getBoundingClientRect();
    this.addContextComponent = new AddDropdown({
      target: document.body,
      props: {
        left: rect.left,
        bottom: window.innerHeight - rect.top,
        onAdd: this.handleContextEnded,
      },
    });
    this.addContextNode = anchorNode;
    this.addContextChar = "@";
  }

  private handleContextEnded = (chatCtx: ChatContextEntry) => {
    if (!this.addContextNode?.parentNode) return;

    const comp = new ChatContext({
      target: this.addContextNode.parentNode,
      anchor: this.addContextNode,
      props: {
        chatCtx,
        metadata: get(this.contextMetadataStore),
        onUpdate: this.updateValue,
      },
    });

    // Wait a loop to ensure the component is added to the DOM.
    setTimeout(() => {
      this.componentAdded(comp, this.addContextNode);

      const selection = window.getSelection();
      const range = selection?.getRangeAt(0);
      if (selection && range) {
        range.setStartBefore(this.addContextNode);
        range.setEndBefore(this.addContextNode);
        selection.removeAllRanges();
        selection.addRange(range);
      }

      this.removeAddContextComponent();
    });
  };

  private removeAddContextComponent(keepTextNode = false) {
    this.addContextComponent?.$destroy();
    this.isContextMode = false;
    if (keepTextNode) return;

    // Remove the search text after context char.
    this.addContextNode.textContent = this.addContextNode.textContent.replace(
      new RegExp(`${this.addContextChar}.*$`),
      "",
    );
    // Only remove the text node if it's empty after removing the search text.
    // If the search was started in the middle of a line, there will be other text.
    if (this.addContextNode.textContent.length === 0)
      this.addContextNode.remove();
  }

  private removeNodes() {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return;

    const range = selection.getRangeAt(0);
    let node: Node | null = range.startContainer;
    if (!node || selection.isCollapsed) return;

    do {
      if (node === this.addContextNode) {
        this.addContextComponent.$destroy();
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
