module.exports = {
  content: [
    "./**/*.html",
    "./**/*.svelte",
    "./src/lib/duckdb-data-types.ts",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
  safelist: [
    'text-blue-800', // needed for blue text in filter pills
  ]
};
