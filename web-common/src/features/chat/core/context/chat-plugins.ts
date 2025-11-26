import Mention, {
  type MentionNodeAttrs,
  type MentionOptions,
} from "@tiptap/extension-mention";
import { Extension } from "@tiptap/core";
import InlineChatContextPicker from "@rilldata/web-common/features/chat/core/context/InlineChatContextPicker.svelte";
import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import InlineChatContextComponent from "@rilldata/web-common/features/chat/core/context/InlineChatContext.svelte";
import {
  convertContextToInlinePrompt,
  parseInlineAttr,
} from "@rilldata/web-common/features/chat/core/context/convertors.ts";
import type { EditorView } from "@tiptap/pm/view";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { Placeholder, UndoRedo } from "@tiptap/extensions";
import {
  INLINE_CHAT_CONTEXT_TAG,
  type InlineChatContext,
  normalizeInlineChatContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export function getEditorPlugins(
  placeholder: string,
  conversationManager: ConversationManager,
  onSubmit: () => void,
) {
  const sharedEditorStore = new SharedEditorStore();

  return [
    Document,
    Paragraph,
    Text,
    Placeholder.configure({
      placeholder,
    }),
    configureInlineContextTipTapExtension(
      conversationManager,
      sharedEditorStore,
    ),
    EditorSubmitExtension.configure({ onSubmit, sharedEditorStore }),
    UndoRedo,
  ];
}

/**
 * Hooks into the editor's shortcut system.
 * Maps Shift-Enter to the editor's enter command.
 * Maps Enter to the submit action calling the onSubmit callback.
 * Also suppresses up and down arrow keys when context picker is open.
 */
const EditorSubmitExtension = Extension.create(() => {
  let isShiftEnter = false;

  return {
    name: "editorSubmit",

    addOptions() {
      return {
        onSubmit: () => {},
        sharedEditorStore: <SharedEditorStore>{},
      };
    },

    addKeyboardShortcuts() {
      return {
        Enter: () => {
          if (!isShiftEnter) {
            // Suppress enter to submit when context picker is open
            if (!this.options.sharedEditorStore.contextOpen) {
              this.options.onSubmit?.();
            }
            return true;
          }
          isShiftEnter = false;
          return false;
        },
        "Shift-Enter": () => {
          isShiftEnter = true;
          return this.editor.commands.enter();
        },
        // Suppress up and down when context picker is open
        ArrowDown: () => {
          return this.options.sharedEditorStore.contextOpen;
        },
        ArrowUp: () => {
          return this.options.sharedEditorStore.contextOpen;
        },
      };
    },
  };
});

/**
 * Extends the existing Mention extension to support inline chat context.
 * Creates InlineChatContext svelte component to display an interactive inline chat context block.
 */
const InlineContextExtension = Mention.extend({
  // Add a param for ConversationManager on top of Mention's options.
  addOptions() {
    return {
      ...this.parent?.(),
      manager: null as ConversationManager | null,
      sharedEditorStore: <SharedEditorStore>{},
    };
  },

  // Mapping for attributes. We need to map values in InlineChatContext to html attribute and vice-versa.
  addAttributes() {
    return {
      type: createAttributeEntry(null, "type"),
      metricsView: createAttributeEntry(null, "metricsView"),
      measure: createAttributeEntry(null, "measure"),
      dimension: createAttributeEntry(null, "dimension"),
      timeRange: createAttributeEntry(null, "timeRange"),
    };
  },

  parseHTML() {
    return [
      {
        tag: INLINE_CHAT_CONTEXT_TAG,
      },
    ];
  },

  renderHTML({ HTMLAttributes }) {
    return [INLINE_CHAT_CONTEXT_TAG, HTMLAttributes, ""];
  },

  renderText({ node }) {
    return convertContextToInlinePrompt(node.attrs as InlineChatContext);
  },

  addNodeView() {
    return ({ node, getPos, view, editor }) => {
      // Create a wrapper div to render the component.
      // We need this since svelte only takes a target wrapper.
      const target = document.createElement("div");
      // We need this here to make sure the component is rendered inline.
      target.className = "inline-block";

      // TODO: fix type so that InlineContextExtension has manager in options
      const sharedEditorStore = (this.options as any)
        .sharedEditorStore as SharedEditorStore;

      // Create the inline chat context component. Pass the wrapper as the target.
      const comp = new InlineChatContextComponent({
        target,
        props: {
          // TODO: fix type so that InlineContextExtension has manager in options
          conversationManager: (this.options as any).manager,
          selectedChatContext: normalizeInlineChatContext(
            node.attrs as InlineChatContext,
          ),
          onSelect: (selectedChatContext) => {
            const pos = getPos();
            if (!pos) return;

            // Dispatch a transaction to update the node attributes with the new context.
            view.dispatch(
              getTransactionForContext(selectedChatContext, view, pos),
            );
            editor.commands.focus();
          },
          onDropdownToggle: (isOpen) =>
            sharedEditorStore.dropdownToggled(comp, isOpen),
          focusEditor: () => editor.commands.focus(),
        },
      });
      sharedEditorStore.componentAdded(comp);

      return {
        dom: target,
        destroy() {
          sharedEditorStore.componentsRemoved(comp);
          comp.$destroy();
        },
      };
    };
  },
});

/**
 * Configures the InlineContextExtension to show a dropdown when the user types "@".
 * Renders the InlineChatContextPicker svelte component.
 */
export function configureInlineContextTipTapExtension(
  manager: ConversationManager,
  sharedEditorStore: SharedEditorStore,
) {
  let comp: InlineChatContextPicker | null = null;

  return InlineContextExtension.configure(<Partial<MentionOptions>>{
    manager,
    sharedEditorStore,
    suggestion: {
      char: "@",
      allowSpaces: true,
      items: () => [],
      render: () => ({
        onStart: (props) => {
          const rect = props.clientRect?.();
          const left = rect?.left ?? 0;
          const bottom = window.innerHeight - (rect?.bottom ?? 0) + 16;

          comp = new InlineChatContextPicker({
            target: document.body,
            props: {
              conversationManager: manager,
              left,
              bottom,
              onSelect: (item) => {
                props.command(item as unknown as MentionNodeAttrs);
              },
              focusEditor: () => props.editor.commands.focus(),
            },
          });
          sharedEditorStore.contextOpen = true;
        },

        onUpdate(props) {
          comp?.$set({ searchText: props.query });
        },

        onExit: ({ editor, range }) => {
          if (!comp) return;
          comp.$destroy();
          comp = null;
          sharedEditorStore.contextOpen = false;

          // Remove the query text and replace with space.
          // This is not automatically removed by tiptap
          editor.view.dispatch(
            editor.view.state.tr.replaceRangeWith(
              range.from + 1,
              range.to + 1,
              editor.state.schema.text(" "),
            ),
          );
        },
      }),
    },
  });
}

/**
 * Used to share data across plugins of editor.
 * It is used to keep track of the state of the context picker dropdowns.
 * It also keeps track of the components that are currently rendered and makes sure only one dropdown is open at a time.
 */
class SharedEditorStore {
  public contextOpen: boolean = false;
  private components: InlineChatContextComponent[] = [];

  public componentAdded(comp: InlineChatContextComponent) {
    this.components.push(comp);
  }

  public componentsRemoved(comp: InlineChatContextComponent) {
    this.components = this.components.filter((c) => c !== comp);
    this.contextOpen = false;
  }

  public dropdownToggled(comp: InlineChatContextComponent, isOpen: boolean) {
    this.contextOpen = isOpen;
    if (!isOpen) return;

    // If the dropdown for the current component was opened, close dropdowns for all other components.
    this.components.forEach((c) => {
      if (c === comp) return;
      c.closeDropdown();
    });
  }
}

function getTransactionForContext(
  inlineChatContext: InlineChatContext,
  view: EditorView,
  pos: number,
) {
  let tr = view.state.tr.setNodeAttribute(pos, "type", inlineChatContext.type);
  if (inlineChatContext.metricsView)
    tr = tr.setNodeAttribute(pos, "metricsView", inlineChatContext.metricsView);
  if (inlineChatContext.measure)
    tr = tr.setNodeAttribute(pos, "measure", inlineChatContext.measure);
  if (inlineChatContext.dimension)
    tr = tr.setNodeAttribute(pos, "dimension", inlineChatContext.dimension);
  if (inlineChatContext.timeRange)
    tr = tr.setNodeAttribute(pos, "timeRange", inlineChatContext.timeRange);
  return tr;
}

function createAttributeEntry(defaultValue: string | null, key: string) {
  return {
    default: defaultValue,
    parseHTML: (element: HTMLElement) =>
      element.getAttribute(key) ?? // Parsing from html attribute.
      parseInlineAttr(element.innerHTML, key) ?? // Parsing from inline prompt.
      defaultValue,
  };
}
