import {
  convertContextToInlinePrompt,
  convertPromptValueToContext,
  INLINE_CHAT_CONTEXT_TAG,
  type InlineChatContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import InlineChatContextComponent from "@rilldata/web-common/features/chat/core/context/InlineChatContext.svelte";
import { getExploreNameStore } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { get } from "svelte/store";
import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import { debounce } from "@rilldata/web-common/lib/create-debouncer.ts";

export class ChatInputTextAreaManager {
  private editorElement: HTMLDivElement;
  private conversationManager: ConversationManager;

  private isContextMode = false;
  private addContextComponent: InlineChatContextComponent | null = null;
  private addContextNode: Node | null = null;
  private elementToContextComponent = new Map<
    Node,
    InlineChatContextComponent
  >();

  private readonly exploreNameStore = getExploreNameStore();

  public constructor(
    private readonly onChange: (newValue: string) => void,
    private readonly onSubmit: () => void,
  ) {}

  public setElement(editorElement: HTMLDivElement) {
    this.editorElement = editorElement;
  }

  public setConversationManager(conversationManager: ConversationManager) {
    this.conversationManager = conversationManager;
  }

  public setPrompt(html: string) {
    this.editorElement.innerHTML = html;

    // Wait till the next tick to ensure DOM is updated after innerHTML change.
    setTimeout(() => {
      // Cleanup any old components
      this.elementToContextComponent.values().forEach((c) => c.$destroy());

      const inlineContextNodes = this.findInlineContextNodes(
        this.editorElement,
      );
      inlineContextNodes.forEach(([parent, node, inlineChatContext]) => {
        const comp = new InlineChatContextComponent({
          target: parent as any,
          anchor: node as any,
          props: {
            conversationManager: this.conversationManager,
            inlineChatContext,
            onUpdate: this.contextUpdated,
            focusEditor: this.focusEditor,
          },
        });
        this.elementToContextComponent.set(node, comp);
        parent.removeChild(node);
      });

      this.focusEditor(true);

      this.updateValue();
    });
  }

  public handleKeydown = (event: KeyboardEvent) => {
    if (!get(this.exploreNameStore)) return; // Only supported within an explore right now.

    const isEnter = event.key === "Enter" && !event.shiftKey;

    if (this.isContextMode) {
      if (event.key === "Escape") {
        this.exitContextMode(false, true);
      } else if (isEnter) {
        this.addContextComponent?.selectFirst();
      } else {
        // Wait for DOM to update.
        setTimeout(() => {
          this.handleContextMode();
        });
      }
      return;
    }

    if (event.key === "Backspace") {
      this.removeNodes();
    } else if (event.key === "@") {
      // Wait for DOM to update.
      setTimeout(() => {
        this.handleContextStarted();
      });
    } else if (isEnter) {
      event.preventDefault();
      this.onSubmit();
    }
  };

  // Add a debounce to avoid triggering the onChange callback on every keystroke.
  public updateValue = debounce(() => {
    let value = this.getValue(this.editorElement);
    if (value === "\n") value = "";
    this.onChange(value);
  }, 100);

  public focusEditor = (setToEnd = false) => {
    this.editorElement.focus();

    if (!setToEnd || !this.editorElement.lastChild) return;

    // Set the cursor to the end of the editor.
    const selection = window.getSelection();
    if (!selection) return;

    const range = selection.getRangeAt(0);
    if (!range) return;

    range.setStartAfter(this.editorElement.lastChild);
    range.setEndAfter(this.editorElement.lastChild);
  };

  private contextUpdated = () => {
    if (this.isContextMode) {
      this.exitContextMode(true, false);
    }

    this.updateValue();
  };

  private getValue(node: Node, level: number = 0): string {
    if (node.nodeType === Node.TEXT_NODE) {
      return node.textContent ?? "";
    } else if (node.nodeType === Node.ELEMENT_NODE) {
      const comp = this.elementToContextComponent.get(node);
      if (comp) {
        const ctx = comp.getChatContext();
        return ctx ? convertContextToInlinePrompt(ctx) : "";
      }

      if (node.nodeName === "BR") return "\n";

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

  private findInlineContextNodes(
    node: Node,
  ): [Node, Node, InlineChatContext | null][] {
    const inlineContextNodes: [Node, Node, InlineChatContext | null][] = [];

    for (const childNode of node.childNodes) {
      if (childNode.nodeName === "DIV") {
        inlineContextNodes.push(...this.findInlineContextNodes(childNode));
        continue;
      }

      if (childNode.nodeName.toLowerCase() !== INLINE_CHAT_CONTEXT_TAG) {
        continue;
      }

      const chatCtx = convertPromptValueToContext(
        (childNode as HTMLElement).innerText,
      );
      if (!chatCtx) continue;

      inlineContextNodes.push([node, childNode, chatCtx]);
    }

    return inlineContextNodes;
  }

  private handleContextMode() {
    if (!this.addContextNode || !this.addContextComponent) return;

    const contextNodeValue: string = this.getValue(this.addContextNode).trim();
    // If the context node is empty then exit context mode, remove the node and component.
    // This happens when the `@` is removed is some way or another.
    if (contextNodeValue.length === 0) {
      this.exitContextMode(true, true);
    } else {
      const searchText = contextNodeValue.replace(/.*?@/, "");
      this.addContextComponent.setText(searchText);
    }
  }

  private handleContextStarted() {
    const selection = window.getSelection();
    if (!selection) return;
    const range = selection.getRangeAt(0);

    const anchorNode = range.startContainer;
    if (!anchorNode) return;

    // Remove the `@` character from the previous node.
    range.setStart(anchorNode, range.startOffset - 1);
    range.deleteContents();

    // Insert a new text node with the `@` character.
    const contextNode = document.createTextNode("@");
    range.insertNode(contextNode);

    this.addContextComponent = new InlineChatContextComponent({
      target: contextNode.parentNode as any,
      anchor: contextNode as any,
      props: {
        conversationManager: this.conversationManager,
        inlineChatContext: null,
        onUpdate: this.contextUpdated,
        focusEditor: this.focusEditor,
      },
    });
    this.isContextMode = true;
    this.addContextNode = contextNode;

    // Wait a loop to ensure the component is added to the DOM.
    setTimeout(() => {
      this.elementToContextComponent.set(
        this.addContextNode!,
        this.addContextComponent!,
      );

      // Move the cursor to the end of search text node.
      range.setStartAfter(contextNode);
      range.setEndAfter(contextNode);
      selection.removeAllRanges();
      selection.addRange(range);
    });
  }

  private removeNodes() {
    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) return;

    const range = selection.getRangeAt(0);
    let node: Node | null = range.startContainer;
    if (!node || selection.isCollapsed) return;

    do {
      const comp = this.elementToContextComponent.get(node);
      comp?.$destroy();
      this.elementToContextComponent.delete(node);

      if (node === this.addContextNode) {
        this.exitContextMode(false, false);
      }

      node = node.nextSibling;
    } while (node && node !== range.endContainer.nextSibling);
  }

  private exitContextMode(
    removeContextNode: boolean,
    removeContextComponent: boolean,
  ) {
    this.isContextMode = false;
    if (removeContextNode) (this.addContextNode as Element)?.remove();
    this.addContextNode = null;
    if (removeContextComponent) this.addContextComponent?.$destroy();
    this.addContextComponent = null;
  }
}
