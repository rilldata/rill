/**
 * ESLint rule to disallow certain Tailwind text color classes.
 * Disallows: text-gray-*, text-neutral-*, text-slate-*, text-stone-*, text-zinc-*
 */

const DISALLOWED_PATTERN = /\btext-(gray|neutral|slate|stone|zinc)-\d{1,3}\b/g;
const ERROR_MESSAGE =
  'Disallowed Tailwind text color class: "{{ className }}". Use semantic color classes instead.';

function reportAllMatches(value, context, node) {
  if (typeof value !== "string") return;

  for (const match of value.matchAll(DISALLOWED_PATTERN)) {
    context.report({
      node,
      message: ERROR_MESSAGE,
      data: { className: match[0] },
    });
  }
}

export default {
  meta: {
    type: "problem",
    docs: {
      description:
        "Disallow non-semantic Tailwind text color classes (gray, neutral, slate, stone, zinc)",
    },
    schema: [],
  },
  create(context) {
    const sourceCode = context.sourceCode ?? context.getSourceCode();

    return {
      // Check string literals in JS/TS
      Literal(node) {
        reportAllMatches(node.value, context, node);
      },
      // Check template literals
      TemplateElement(node) {
        reportAllMatches(node.value.raw, context, node);
      },
      // Check Svelte HTML attributes (class="...")
      SvelteAttribute(node) {
        if (node.key?.name === "class") {
          for (const valueNode of node.value) {
            if (valueNode.type === "SvelteLiteral") {
              reportAllMatches(valueNode.value, context, valueNode);
            }
          }
        }
      },
      // Check Svelte shorthand class directives (class:text-gray-500)
      SvelteDirective(node) {
        if (node.kind === "Class" && node.key?.name) {
          const className = node.key.name.name || node.key.name;
          reportAllMatches(className, context, node);
        }
      },
      // Check Svelte <style> blocks
      SvelteStyleElement(node) {
        const styleText = sourceCode.getText(node);
        const nodeStart = node.range[0];

        for (const match of styleText.matchAll(DISALLOWED_PATTERN)) {
          const matchStart = nodeStart + match.index;
          const matchEnd = matchStart + match[0].length;

          context.report({
            loc: {
              start: sourceCode.getLocFromIndex(matchStart),
              end: sourceCode.getLocFromIndex(matchEnd),
            },
            message: ERROR_MESSAGE,
            data: { className: match[0] },
          });
        }
      },
    };
  },
};
