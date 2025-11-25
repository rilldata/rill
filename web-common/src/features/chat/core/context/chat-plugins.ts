import Mention, {
  type MentionNodeAttrs,
  type MentionOptions,
} from "@tiptap/extension-mention";
import { Extension, mergeAttributes } from "@tiptap/core";
import AddInlineChatDropdown from "@rilldata/web-common/features/chat/core/context/AddInlineChatDropdown.svelte";
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
  INLINE_CHAT_DIMENSION_ATTR,
  INLINE_CHAT_MEASURE_ATTR,
  INLINE_CHAT_METRICS_VIEW_ATTR,
  INLINE_CHAT_TIME_RANGE_ATTR,
  INLINE_CHAT_TYPE_ATTR,
  type InlineChatContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export function getEditorPlugins(
  placeholder: string,
  conversationManager: ConversationManager,
  onSubmit: () => void,
) {
  return [
    Document,
    Paragraph,
    Text,
    Placeholder.configure({
      placeholder,
    }),
    configureInlineContextTipTapExtension(conversationManager),
    UndoRedo,
    EditorSubmitExtension.configure({ onSubmit }),
  ];
}

/**
 * Hooks into the editor's shortcut system.
 * Maps Shift-Enter to the editor's enter command.
 * Maps Enter to the submit action calling the onSubmit callback.
 */
const EditorSubmitExtension = Extension.create(() => {
  let isShiftEnter = false;

  return {
    name: "editorSubmit",

    addOptions() {
      return {
        onSubmit: () => {},
      };
    },

    addKeyboardShortcuts() {
      return {
        Enter: () => {
          if (!isShiftEnter) {
            this.options.onSubmit?.();
          }
          isShiftEnter = false;
          return true;
        },
        "Shift-Enter": () => {
          isShiftEnter = true;
          return this.editor.commands.enter();
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
    };
  },

  // Mapping for attributes. We need to map values in InlineChatContext to html attribute and vice-versa.
  addAttributes() {
    return {
      type: createAttributeEntry(null, "type", INLINE_CHAT_TYPE_ATTR),
      metricsView: createAttributeEntry(
        null,
        "metricsView",
        INLINE_CHAT_METRICS_VIEW_ATTR,
      ),
      measure: createAttributeEntry(null, "measure", INLINE_CHAT_MEASURE_ATTR),
      dimension: createAttributeEntry(
        null,
        "dimension",
        INLINE_CHAT_DIMENSION_ATTR,
      ),
      timeRange: createAttributeEntry(
        null,
        "timeRange",
        INLINE_CHAT_TIME_RANGE_ATTR,
      ),
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
    return [INLINE_CHAT_CONTEXT_TAG, HTMLAttributes];
  },

  renderText({ node }) {
    return convertContextToInlinePrompt(node.attrs as InlineChatContext);
  },

  addNodeView() {
    return ({ node, getPos, view }) => {
      // Create a wrapper div to render the component.
      // We need this since svelte only takes a target wrapper.
      const target = document.createElement("div");
      // We need this here to make sure the component is rendered inline.
      target.className = "inline-block";

      view.dispatch(view.state.tr.deleteSelection());

      // Create the inline chat context component. Pass the wrapper as the target.
      const comp = new InlineChatContextComponent({
        target,
        props: {
          // TODO: fix type so that InlineContextExtension has manager in options
          conversationManager: (this.options as any).manager,
          inlineChatContext: node.attrs as InlineChatContext,
          onSelect: (inlineChatContext) => {
            const pos = getPos();
            if (!pos) return;

            // Dispatch a transaction to update the node attributes with the new context.
            view.dispatch(
              getTransactionForContext(inlineChatContext, view, pos),
            );
          },
        },
      });

      return {
        dom: target,
        destroy() {
          comp.$destroy();
        },
      };
    };
  },
});

/**
 * Configures the InlineContextExtension to show a dropdown when the user types "@".
 * Renders the AddInlineChatDropdown svelte component.
 */
export function configureInlineContextTipTapExtension(
  manager: ConversationManager,
) {
  let comp: AddInlineChatDropdown | null = null;

  return InlineContextExtension.configure(<Partial<MentionOptions>>{
    manager,
    deleteTriggerWithBackspace: true,
    suggestion: {
      char: "@",
      items: () => [],
      render: () => ({
        onStart: (props) => {
          const rect = props.clientRect?.();
          const left = rect?.left ?? 0;
          const bottom = window.innerHeight - (rect?.bottom ?? 0) + 16;

          comp = new AddInlineChatDropdown({
            target: document.body,
            props: {
              conversationManager: manager,
              left,
              bottom,
              onSelect: (item) => {
                props.command(item as unknown as MentionNodeAttrs);
              },
            },
          });
        },
        onUpdate(props) {
          comp?.$set({ searchText: props.text });
        },
        onExit: () => {
          comp?.$destroy();
          comp = null;
        },
      }),
    },
  });
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

function createAttributeEntry(
  defaultValue: string | null,
  key: string,
  attr: string,
) {
  return {
    default: defaultValue,
    parseHTML: (element: HTMLElement) =>
      element.getAttribute(attr) ?? // Parsing from html attribute.
      parseInlineAttr(element.innerHTML, key) ?? // Parsing from inline prompt.
      defaultValue,
    renderHTML: (attributes: Record<string, string>) =>
      mergeAttributes(attributes, { [attr]: attributes[key] }),
  };
}
