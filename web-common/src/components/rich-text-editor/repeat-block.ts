import { Node, mergeAttributes } from "@tiptap/core";

export interface RepeatBlockOptions {
  HTMLAttributes: Record<string, any>;
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    repeatBlock: {
      setRepeatBlock: (options: {
        measure: string;
        dimension: string;
        orderBy?: string;
        limit?: number;
        where?: string;
      }) => ReturnType;
    };
  }
}

export const RepeatBlock = Node.create<RepeatBlockOptions>({
  name: "repeatBlock",

  addOptions() {
    return {
      HTMLAttributes: {},
    };
  },

  group: "block",

  content: "block+",

  addAttributes() {
    return {
      measure: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-measure"),
        renderHTML: (attributes) => {
          if (!attributes.measure) {
            return {};
          }
          return {
            "data-measure": attributes.measure,
          };
        },
      },
      dimension: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-dimension"),
        renderHTML: (attributes) => {
          if (!attributes.dimension) {
            return {};
          }
          return {
            "data-dimension": attributes.dimension,
          };
        },
      },
      orderBy: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-order-by"),
        renderHTML: (attributes) => {
          if (!attributes.orderBy) {
            return {};
          }
          return {
            "data-order-by": attributes.orderBy,
          };
        },
      },
      limit: {
        default: null,
        parseHTML: (element) => {
          const limit = element.getAttribute("data-limit");
          return limit ? parseInt(limit, 10) : null;
        },
        renderHTML: (attributes) => {
          if (!attributes.limit) {
            return {};
          }
          return {
            "data-limit": attributes.limit.toString(),
          };
        },
      },
      where: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-where"),
        renderHTML: (attributes) => {
          if (!attributes.where) {
            return {};
          }
          return {
            "data-where": attributes.where,
          };
        },
      },
    };
  },

  parseHTML() {
    return [
      {
        tag: 'div[data-type="repeat-block"]',
      },
    ];
  },

  renderHTML({ node, HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(this.options.HTMLAttributes, HTMLAttributes, {
        "data-type": "repeat-block",
        class: "repeat-block",
      }),
      0,
    ];
  },

  addCommands() {
    return {
      setRepeatBlock:
        (options) =>
        ({ commands }) => {
          return commands.insertContent({
            type: this.name,
            attrs: options,
            content: [
              {
                type: "paragraph",
              },
            ],
          });
        },
    };
  },
});

