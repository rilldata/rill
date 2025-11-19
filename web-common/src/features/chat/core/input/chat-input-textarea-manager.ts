import AddDropdown from "@rilldata/web-common/features/chat/core/context/AddDropdown.svelte";
import AddValueDropdown from "@rilldata/web-common/features/chat/core/context/AddValueDropdown.svelte";
import ChatContext from "@rilldata/web-common/features/chat/core/context/ChatContext.svelte";
import {
  type ChatContextEntry,
  ChatContextEntryType,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import {
  convertContextToInlinePrompt,
  convertHTMLElementToContext,
} from "@rilldata/web-common/features/chat/core/context/conversions.ts";
import { getContextMetadataStore } from "@rilldata/web-common/features/chat/core/context/get-context-metadata-store.ts";
import { get } from "svelte/store";

const SPACE_TEXT = "\u00A0";

export class ChatInputTextAreaManager {
  private editorElement: HTMLDivElement;
  private onChange: (newValue: string) => void;
  private onSubmit: () => void;

  private isContextMode = false;
  private addContextComponent;
  private addContextNode;
  private addContextChar: string = "";
  private elementToContextComponent = new Map<Node, ChatContext>();

  private readonly contextMetadataStore = getContextMetadataStore();

  public setElement(editorElement: HTMLDivElement) {
    this.editorElement = editorElement;
  }

  public setOnChange(onChange: (newValue: string) => void) {
    this.onChange = onChange;
  }

  public setOnSubmit(onSubmit: () => void) {
    this.onSubmit = onSubmit;
  }

  public setHtml(html: string) {
    this.editorElement.innerHTML = html;
    setTimeout(() => {
      this.editorElement.focus();
      this.elementToContextComponent.values().forEach((c) => c.$destroy());
      this.findInlineContextNodes(this.editorElement).forEach(
        ([parent, node, chatCtx]) => {
          const comp = new ChatContext({
            target: parent as any,
            anchor: node as any,
            props: {
              chatCtx,
              metadata: get(this.contextMetadataStore),
              onUpdate: () => this.updateValue(),
            },
          });
          this.componentAdded(comp, node);
          parent.removeChild(node);
        },
      );
    });
  }

  public handleKeydown = (event: KeyboardEvent) => {
    setTimeout(() => {
      if (event.key === "Backspace") {
        this.removeNodes();
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
      const ctx = convertHTMLElementToContext(
        node as any,
        get(this.contextMetadataStore),
      );
      if (ctx) return convertContextToInlinePrompt(ctx);

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

      const chatCtx = convertHTMLElementToContext(
        childNode as HTMLElement,
        get(this.contextMetadataStore),
      );
      if (!chatCtx) continue;

      inlineContextNodes.push([node, childNode, chatCtx]);
    }

    return inlineContextNodes;
  }

  private isChildOfEditor(node: Node | null | undefined) {
    if (!node) return false;
    if (this.editorElement.contains(node)) return true;
    return this.isChildOfEditor(node.parentNode);
  }

  private insertSvelteComponent(
    componentCreator: (
      target: Element | Document | ShadowRoot,
      anchor: Element | undefined,
    ) => Node | null | undefined,
  ) {
    const selection = window.getSelection();

    if (
      selection &&
      selection.rangeCount > 0 &&
      this.isChildOfEditor(selection.anchorNode?.parentNode)
    ) {
      const range = selection.getRangeAt(0);
      range.deleteContents();

      const node = componentCreator(
        selection.anchorNode?.parentNode as any,
        selection.anchorNode as any,
      );
      if (node) {
        // Move cursor after the node
        range.setStartAfter(node);
        range.setEndAfter(node);
      }

      selection.removeAllRanges();
      selection.addRange(range);
    } else {
      const node = componentCreator(this.editorElement, undefined);
      if (node) {
        if (this.editorElement.childNodes.length === 0) {
          this.editorElement.appendChild(node);
        } else {
          this.editorElement.insertBefore(node, this.editorElement.firstChild);
        }
      }
    }

    this.updateValue();
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
    if (
      range.startOffset !== 1 ||
      !(range.startContainer as any)?.previousElementSibling
    )
      return;

    const prevSibling = (range.startContainer as any)
      ?.previousElementSibling as HTMLElement;
    const comp = this.elementToContextComponent.get(prevSibling);
    if (!comp) return;

    const ctx = convertHTMLElementToContext(
      prevSibling,
      get(this.contextMetadataStore),
    );
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

    this.insertSvelteComponent(() => {
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
      return anchorNode;
    });
  }

  private handleContextEnded = (chatCtx: ChatContextEntry) => {
    if (!this.addContextNode?.parentNode) return;
    const spaceNode = document.createTextNode(SPACE_TEXT);
    if (this.addContextNode.nextSibling) {
      this.addContextNode.parentNode.insertBefore(
        spaceNode,
        this.addContextNode.nextSibling,
      );
    } else {
      this.addContextNode.parentNode.appendChild(spaceNode);
    }

    const comp = new ChatContext({
      target: this.addContextNode.parentNode,
      anchor: spaceNode as any,
      props: {
        chatCtx,
        metadata: get(this.contextMetadataStore),
        onUpdate: () => this.updateValue(),
      },
    });
    this.componentAdded(comp, spaceNode);
    this.removeAddContextComponent();

    const selection = window.getSelection();
    const range = selection?.getRangeAt(0);
    if (selection && range) {
      range.setStartBefore(spaceNode);
      range.setEndBefore(spaceNode);
      selection.removeAllRanges();
      selection.addRange(range);
    }
    spaceNode.remove();
  };

  private removeAddContextComponent() {
    this.addContextComponent?.$destroy();
    this.isContextMode = false;
    this.addContextNode.textContent = this.addContextNode.textContent.replace(
      new RegExp(`${this.addContextChar}.*$`),
      "",
    );
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
