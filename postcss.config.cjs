// postcss.config.js

module.exports = {
  plugins: {
    autoprefixer: {},
    tailwindcss: {
      content: ["./src/**/*.{html,js}"],
    },
  },
};
