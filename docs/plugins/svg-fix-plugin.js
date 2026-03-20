/**
 * Docusaurus plugin that configures @svgr/webpack to prefix SVG IDs,
 * preventing collisions when multiple inline SVGs share the same page.
 *
 * Without this, svgo's cleanupIDs plugin renames all IDs to short sequential
 * names (a, b, c…), causing the browser to resolve url(#a) to the first
 * definition in DOM order — typically Supabase's green gradient.
 */
module.exports = function svgFixPlugin() {
  return {
    name: 'svg-fix-plugin',
    configureWebpack(config) {
      // Find the rule that handles SVGs via @svgr/webpack
      const rules = config.module?.rules ?? [];
      for (const rule of rules) {
        if (!rule?.oneOf) continue;
        for (const oneOfRule of rule.oneOf) {
          if (!oneOfRule?.use) continue;
          const uses = Array.isArray(oneOfRule.use) ? oneOfRule.use : [oneOfRule.use];
          for (const use of uses) {
            const loader = typeof use === 'string' ? use : use?.loader;
            if (loader && loader.includes('@svgr/webpack')) {
              const options = typeof use === 'object' ? use : {};
              if (!options.options) options.options = {};
              options.options.svgo = true;
              options.options.svgoConfig = {
                plugins: [
                  {
                    name: 'prefixIds',
                    params: {
                      prefixIds: true,
                      prefixClassNames: true,
                    },
                  },
                ],
              };
            }
          }
        }
      }
      return {};
    },
  };
};
