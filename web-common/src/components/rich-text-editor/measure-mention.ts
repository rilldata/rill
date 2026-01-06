import Mention, { type MentionOptions } from "@tiptap/extension-mention";

export interface MeasureMentionOptions extends MentionOptions {
  metricsViewName: string;
  availableMeasures: string[];
}

export const MeasureMention = Mention.extend<MeasureMentionOptions>({
  name: "measureMention",

  addOptions() {
    return {
      ...this.parent?.(),
      metricsViewName: "",
      availableMeasures: [],
      HTMLAttributes: {
        class: "measure-mention",
        "data-type": "measure",
      },
    };
  },

  addAttributes() {
    return {
      ...this.parent?.(),
      id: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-id"),
        renderHTML: (attributes) => {
          if (!attributes.id) {
            return {};
          }
          return {
            "data-id": attributes.id,
          };
        },
      },
      label: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-label"),
        renderHTML: (attributes) => {
          if (!attributes.label) {
            return {};
          }
          return {
            "data-label": attributes.label,
          };
        },
      },
    };
  },

  renderHTML({ node, HTMLAttributes }) {
    return [
      "span",
      {
        ...this.options.HTMLAttributes,
        ...HTMLAttributes,
        "data-measure": node.attrs.id,
        "data-metrics-view": this.options.metricsViewName,
      },
      `@${node.attrs.label || node.attrs.id}`,
    ];
  },

  renderText({ node }) {
    return `@${node.attrs.label || node.attrs.id}`;
  },
});

