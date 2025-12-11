import Mention, { type MentionOptions } from "@tiptap/extension-mention";
import { Extension } from "@tiptap/core";
import InlineContextPicker from "@rilldata/web-common/features/chat/core/context/picker/InlineContextPicker.svelte";
import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import InlineContextComponent from "@rilldata/web-common/features/chat/core/context/InlineContext.svelte";
import {
  convertContextToInlinePrompt,
  parseInlineAttr,
} from "@rilldata/web-common/features/chat/core/context/inline-context-convertors.ts";
import type { EditorView } from "@tiptap/pm/view";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { Placeholder, UndoRedo } from "@tiptap/extensions";
import {
  INLINE_CHAT_CONTEXT_TAG,
  type InlineContext,
  normalizeInlineContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export function getEditorPlugins({
  enableMention,
  placeholder,
  onSubmit,
}: {
  enableMention: boolean;
  placeholder: string;
  onSubmit: () => void;
}) {
  const sharedEditorStore = new SharedEditorStore();

  const plugins = [
    Document,
    Paragraph,
    Text,
    Placeholder.configure({
      placeholder,
    }),
    EditorSubmitExtension.configure({ onSubmit, sharedEditorStore }),
    UndoRedo,
  ];

  if (enableMention) {
    plugins.push(configureInlineContextTipTapExtension(sharedEditorStore));
  }

  return plugins;
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

type InlineContextOptions = MentionOptions<never, InlineContext> & {
  sharedEditorStore: SharedEditorStore;
};

// Add the startMention command to the Commands type.
declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    mention: {
      startMention: () => ReturnType;
    };
  }
}

/**
 * Extends the existing Mention extension to support inline chat context.
 * Creates InlineChatContext svelte component to display an interactive inline chat context block.
 */
const InlineContextExtension = Mention.extend<InlineContextOptions>({
  // Add a param for ConversationManager on top of Mention's options.
  addOptions() {
    return {
      ...((this.parent?.() ?? {}) as MentionOptions<never, InlineContext>),
      // These have to be configured for the extension to work
      manager: {} as ConversationManager,
      sharedEditorStore: {} as SharedEditorStore,
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
      filePath: createAttributeEntry(null, "filePath"),
    };
  },

  addCommands() {
    return {
      startMention:
        () =>
        ({ tr, view, commands }) => {
          commands.focus();
          // Only focus the editor if context is already open.
          if (this.options.sharedEditorStore.contextOpen) return false;

          tr.insertText("@");
          view.dispatchEvent(new KeyboardEvent("keyup", { key: "@" }));
          return true;
        },
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
    return convertContextToInlinePrompt(node.attrs as InlineContext);
  },

  addNodeView() {
    return ({ node, getPos, view, editor }) => {
      // Create a wrapper div to render the component.
      // We need this since svelte only takes a target wrapper.
      const target = document.createElement("div");
      // We need this here to make sure the component is rendered inline.
      target.className = "inline-block";

      const sharedEditorStore = this.options.sharedEditorStore;

      // Create the inline chat context component. Pass the wrapper as the target.
      const comp = new InlineContextComponent({
        target,
        props: {
          selectedChatContext: normalizeInlineContext(
            node.attrs as InlineContext,
          ),
          props: editor.options.editable
            ? {
                mode: "editable",
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
              }
            : { mode: "readonly" },
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
 * Renders the InlineContextPicker svelte component.
 */
export function configureInlineContextTipTapExtension(
  sharedEditorStore: SharedEditorStore,
) {
  let comp: InlineContextPicker | null = null;
  let selected = false;

  return InlineContextExtension.configure({
    sharedEditorStore,
    suggestion: {
      char: "@",
      allowSpaces: true,
      items: () => [], // TODO: would it make sense to manage the options here?
      render: () => ({
        onStart: (props) => {
          if (!(props.decorationNode instanceof HTMLElement)) return; // type safety, non-html will be in non-dom environment
          selected = false;

          comp = new InlineContextPicker({
            target: document.body,
            props: {
              refNode: props.decorationNode,
              onSelect: (item) => {
                selected = true;
                props.command(item);
              },
              focusEditor: () => props.editor.commands.focus(),
            },
          });
          sharedEditorStore.contextOpen = true;
        },

        onUpdate(props) {
          if (!(props.decorationNode instanceof HTMLElement)) return; // type safety, non-html will be in non-dom environment
          comp?.$set({
            searchText: props.query,
            refNode: props.decorationNode,
          });
        },

        onExit: ({ editor, range }) => {
          if (!comp) return;
          comp.$destroy();
          comp = null;
          sharedEditorStore.contextOpen = false;

          if (!selected) return;
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
  private components: InlineContextComponent[] = [];

  public componentAdded(comp: InlineContextComponent) {
    this.components.push(comp);
  }

  public componentsRemoved(comp: InlineContextComponent) {
    this.components = this.components.filter((c) => c !== comp);
    this.contextOpen = false;
  }

  public dropdownToggled(comp: InlineContextComponent, isOpen: boolean) {
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
  inlineChatContext: InlineContext,
  view: EditorView,
  pos: number,
) {
  return view.state.tr
    .setNodeAttribute(pos, "type", inlineChatContext.type)
    .setNodeAttribute(pos, "metricsView", inlineChatContext.metricsView)
    .setNodeAttribute(pos, "measure", inlineChatContext.measure)
    .setNodeAttribute(pos, "dimension", inlineChatContext.dimension)
    .setNodeAttribute(pos, "timeRange", inlineChatContext.timeRange)
    .setNodeAttribute(pos, "filePath", inlineChatContext.filePath);
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
