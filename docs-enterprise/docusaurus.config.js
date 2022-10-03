// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

/* eslint @typescript-eslint/no-var-requires: "off" */
const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

const def = require("redocusaurus");
def;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Rill",
  tagline: "A simple alternative to complex BI stacks",

  // netlify settings
  url: "https://enterprise.rilldata.com",
  baseUrl: "/",

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon.ico",

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: "/",
          sidebarPath: require.resolve("./sidebars.js"),
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            "https://github.com/rilldata/rill-developer/blob/main/docs-enterprise/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // algolia: {
      //   appId: "O4A4YNY97A",
      //   apiKey: "f7bc8583cf7d74049dd1bd937cf42685",
      //   indexName: "rill_enterprise",
      //   debug: false, // Set debug to true if you want to inspect the modal
      // },
      docs: {
        sidebar: {
          autoCollapseCategories: true,
        },
      },
      metadata: [
        {
          property: "og:image",
          content:
            "https://images.ctfassets.net/ve6smfzbifwz/5MvW4kOHMbGBIIAI7hWe65/a9418adf8f96ee0d3a3ca1341f368e67/Rill_Data.png",
        },
        {
          name: "twitter:image",
          content:
            "https://images.ctfassets.net/ve6smfzbifwz/5MvW4kOHMbGBIIAI7hWe65/a9418adf8f96ee0d3a3ca1341f368e67/Rill_Data.png",
        },
      ],
      navbar: {
        logo: {
          alt: "Rill Logo",
          src: "img/rill-logo-light.svg",
          srcDark: "img/rill-logo-dark.svg",
          href: "https://app.rilldata.com",
          target: "_self",
        },
        items: [
          { to: "https://rilldata.com", position: "left", label: "Home" },
          {
            to: "https://rilldata.com/product",
            position: "left",
            label: "Product",
          },
          {
            to: "https://rilldata.com/apache-druid",
            position: "left",
            label: "Apache Druid",
          },
          { to: "https://rilldata.com/team", position: "left", label: "Team" },
          { to: "https://rilldata.com/blog", position: "left", label: "Blog" },
          {
            to: "https://rilldata.com/try-free",
            position: "left",
            label: "Try for Free",
          },
          { type: "search", position: "right" },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            title: " ",
            items: [
              {
                label: "Rill Data",
                to: "https://www.rilldata.com",
              },
              {
                label: "Enterprise Docs",
                to: "/",
              },
            ],
          },
          {
            title: " ",
            items: [
              {
                html: `
                 <div style="display: flex; align-items: center; -webkit-box-align: center;">
                 <a class="social-link" href="https://github.com/rilldata/rill-developer" target="_blank"><img src="https://uploads-ssl.webflow.com/624f2a9ba37f4233dbe55d72/625af1b8081e31a5e696066b_github-octocat.svg" loading="lazy" alt="github logo"></a>
                 <a class="social-link" href="https://twitter.com/RillData" target="_blank"><img src="https://uploads-ssl.webflow.com/624f2a9ba37f4233dbe55d72/624f2a9ba37f429995e55f34_social-twitter.svg" loading="lazy" alt="twitter logo"></a>
                 <a class="social-link" href="https://discord.gg/eEvSYHdfWK" target="_blank"><img src="https://uploads-ssl.webflow.com/624f2a9ba37f4233dbe55d72/625af1dc6a667e2367b552ae_Discord-Logo.svg" loading="lazy" alt="Discord logo"></a>
                 </div>
                `,
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Rill Data, Inc.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ["java"],
      },
    }),

  plugins: [
    [
      require.resolve("docusaurus-gtm-plugin"),
      {
        id: "GTM-TH485ZV",
      },
    ],
  ],
};

module.exports = config;
