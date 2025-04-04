// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

/* eslint @typescript-eslint/no-var-requires: "off" */
const { themes } = require('prism-react-renderer');
const lightCodeTheme = themes.github;
const darkCodeTheme = themes.dracula;

const def = require("redocusaurus");
def;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Rill",
  tagline: "A simple alternative to complex BI stacks",

  // netlify settings
  url: "https://docs.rilldata.com",
  baseUrl: "/",

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "throw",
  favicon: "img/favicon.png",

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
          sidebarCollapsed: true,
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            "https://github.com/rilldata/rill/blob/main/docs/",
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
          customCss: require.resolve("./src/css/custom.scss"),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      algolia: {
        appId: "4U01DM4NS4",
        apiKey: "c0399915ae21a35c6d34a473d017c15b",
        indexName: "rilldata",
        debug: false // Set debug to true if you want to inspect the modal        
      },
      metadata: [
        {
          property: 'og:title', content: "Rill | Fast operational dashboards you'll actually use"
        },
        {
          property: 'og:image', content: 'https://assets-global.website-files.com/659ddac460dbacbdc813b204/65bad0233db92db596c29c34_social1.jpg'
        },
        {
          name: 'twitter:image', content: 'https://assets-global.website-files.com/659ddac460dbacbdc813b204/65bad0233db92db596c29c34_social1.jpg'
        },
        {
          name: 'description', content: "Rill is an operational BI tool that helps data teams build fewer, more flexible dashboards, and helps business users make faster decisions with fewer ad hoc requests."
        }
      ],
      navbar: {
        logo: {
          alt: "Rill Logo",
          src: "img/rill-logo-light.svg",
          srcDark: "img/rill-logo-dark.svg",
          href: "https://www.rilldata.com",
          target: "_self",
        },
        items: [
          {
            type: "doc",
            docId: "home/home",
            position: "left",
            label: "Docs",
          },
          {
            type: "docSidebar",
            sidebarId: "tutorialsSidebar",
            position: "left",
            label: "Tutorials",
          },
          {
            type: "docSidebar",
            sidebarId: "refSidebar",
            position: "left",
            label: "Reference",
          },
         
          {
            label: "Release Notes",
            to: "notes",
            position: "left",
          },

          {
            to: "contact",
            position: "left",
            label: "Contact Us",
          },
          {
            href: "https://github.com/rilldata/rill",
            label: "GitHub",
            position: "right",
          },
          {
            href: "https://www.rilldata.com/blog",
            label: "Blog",
            position: "right",
          },
          {
            type: "search",
            position: "right"
          }
        ],
      },
      footer: {
        style: "light",
        copyright: `© ${new Date().getFullYear()} Rill Data, Inc. • <a href="https://www.rilldata.com/legal/privacy" target="_blank">Privacy Policy</a> • <a href="https://www.rilldata.com/legal/tos" target="_blank"> Terms of Service </a> • <a href="https://github.com/rilldata/rill/blob/main/COMMUNITY-POLICY.md" target="_blank"> Community Policy </a> • <a href="https://github.com/rilldata/rill/blob/main/CONTRIBUTING.md" target="_blank"> Contributing </a>`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ['bash', 'diff', 'json'],
      },
    }),

  plugins: [
    'docusaurus-plugin-sass',
    [
      require.resolve('docusaurus-gtm-plugin'),
      {
        id: 'GTM-TH485ZV',
      }
    ],
    [
      '@docusaurus/plugin-client-redirects',
      {
        redirects: [
          {
            from: '/install',
            to: '/home/install',
          },
          {
            from: '/get-started',
            to: '/home/get-started',
          },
          {
            from: '/develop/import-data',
            to: '/build/connect'
          },
          {
            from: '/develop/sql-models',
            to: '/build/models'
          },
          {
            from: '/develop/metrics-dashboard',
            to: '/build/dashboards'
          },
          {
            from: '/develop/security',
            to: '/manage/security'
          },
          {
            from: '/deploy/credentials/',
            to: '/build/credentials'
          },
          {
            from: '/deploy/credentials/s3',
            to: '/reference/connectors/s3'
          },
          {
            from: '/deploy/credentials/gcs',
            to: '/reference/connectors/gcs'
          },
          {
            from: '/deploy/credentials/azure',
            to: '/reference/connectors/azure'
          },
          {
            from: '/deploy/credentials/athena',
            to: '/reference/connectors/athena'
          },
          {
            from: '/deploy/credentials/bigquery',
            to: '/reference/connectors/bigquery'
          },
          {
            from: '/deploy/credentials/snowflake',
            to: '/reference/connectors/snowflake'
          },
          {
            from: '/deploy/credentials/postgres',
            to: '/reference/connectors/postgres'
          },
          {
            from: '/deploy/credentials/salesforce',
            to: '/reference/connectors/salesforce'
          },
          {
            from: '/deploy/credentials/motherduck',
            to: '/reference/connectors/motherduck'
          },
          {
            from: '/deploy/source-refresh',
            to: '/build/connect/source-refresh'
          },
          {
            from: '/reference/templating',
            to: '/deploy/templating'
          },
          {
            from: '/example-projects',
            to: '/home/get-started#example-projects-repository'
          },
          {
            from: '/integration/embedding',
            to: '/integrate/embedding'
          },
          {
            from: '/share/user-management',
            to: '/manage/user-management'
          },
          {
            from: '/share/roles-permissions',
            to: '/manage/roles-permissions'
          },
          {
            from: '/share/scheduled-reports',
            to: '/explore/exports'
          },
        ],
      },
    ],
  ],

  // Configure Mermaid for diagrams
  themes: ['@docusaurus/theme-mermaid'],
  markdown: {
    mermaid: true,
  },
};

module.exports = config;
