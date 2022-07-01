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
  url: "https://rill-developer.netlify.app",
  baseUrl: "/",
  // gitpages
  // url: "https://rilldata.github.io",
  // baseUrl: "/rill-developer/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon.ico",

  // GitHub pages deployment config.
  // organizationName: "rilldata",
  // projectName: "rill-developer",
  // deploymentBranch: "gh-pages",
  trailingSlash: true,

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
            "https://github.com/rilldata/rill-developer/blob/main/docs/",
        },
        blog: {
          routeBasePath: 'notes',
          blogTitle: 'Release Notes',
          blogDescription: 'Release notes for Rill',
          postsPerPage: 1,
          blogSidebarTitle: 'Release Notes',
          blogSidebarCount: 'ALL',
          feedOptions: {
            type: 'all',
            copyright: `Copyright © ${new Date().getFullYear()} Rill Data, Inc.`,
          },
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
      metadata: [
        { 
          property: 'og:image', content: 'https://uploads-ssl.webflow.com/5e4306d09c892720b9be39a6/607dc6fd92da47780c40b359_Opengraph.png'
        },
        {
          name: 'twitter:image', content: 'https://uploads-ssl.webflow.com/5e4306d09c892720b9be39a6/607dc6fd92da47780c40b359_Opengraph.png'
        },
      ],
      navbar: {
        logo: {
          alt: "Rill Logo",
          src: "img/logo.svg",
          href: "https://www.rilldata.com",
          target: "_self",
        },
        items: [
          {
            type: "doc",
            docId: "README",
            position: "left",
            label: "Docs",
          },
          {
            label: "Release Notes",
            to: "notes",
            position: "left",
          },
          {
            href: "https://github.com/rilldata/rill-developer",
            label: "GitHub",
            position: "left",
          },
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
                label: "Docs",
                to: "/",
              },
              {
                label: "Release Notes",
                to: "/notes",
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
                `
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Rill Data, Inc.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
    }),
  
  plugins: [
    [
      require.resolve('docusaurus-gtm-plugin'),
      {
        id: 'GTM-TH485ZV',
      }
    ]
  ]
};

module.exports = config;
