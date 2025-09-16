// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

/* eslint @typescript-eslint/no-var-requires: "off" */
const { themes } = require('prism-react-renderer');
const lightCodeTheme = themes.github;
const darkCodeTheme = themes.dracula;

const llmsTxtPlugin = require('./plugins/llms-txt-plugin');

const def = require("redocusaurus");
def;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Rill",
  tagline: "A simple alternative to complex BI stacks",

  // netlify settings
  url: "https://docs.rilldata.com",
  baseUrl: "/",
  trailingSlash: false,

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
          sidebarCollapsed: true
        },

        blog: {
          routeBasePath: 'notes',
          blogTitle: 'Release Notes',
          blogDescription: 'Release notes for Rill',
          postsPerPage: 1,
          blogSidebarTitle: 'Release Notes',
          blogSidebarCount: 'ALL',
          onUntruncatedBlogPosts: 'ignore',
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
      colorMode: {
        defaultMode: 'light',
        disableSwitch: false,
        respectPrefersColorScheme: false,
      },
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
          href: "/",
          target: "_self",
        },
        items: [
          {
            type: "doc",
            docId: "get-started/get-started",
            position: "left",
            label: "Docs",
          },

          {
            type: "docSidebar",
            sidebarId: "refSidebar",
            position: "left",
            label: "Reference",
          },

          {
            type: "docSidebar",
            sidebarId: "contactSidebar",
            position: "left",
            label: "Contact Us",
          },

          {
            type: "html",
            position: "right",
            value: '<a href="https://docs.rilldata.com/notes" class="navbar-release-notes-mobile navbar-icon-link" aria-label="Release Notes">Release Notes</i></a>',
          },

          // Right side items
          {
            type: "html",
            position: "right",
            value: '<a href="https://github.com/rilldata/rill" class="navbar-icon-link" aria-label="GitHub">GitHub</i></a>',
          },
          {
            type: "html",
            position: "right",
            value: '<a href="https://www.rilldata.com/blog" class="navbar-icon-link" aria-label="Blog">Blog</i></a>',
          },

          {
            type: "search",
            position: "right"
          },
          {
            type: "html",
            position: "right",
            value: '<span class="navbar-divider"></span>',
          },
          // {
          //   type: "html",
          //   position: "right",
          //   value: '<button id="dark-mode-toggle" class="navbar-icon-link" aria-label="Toggle dark mode"><div class="icon-container"></div></button>',
          // },

          // {
          //   type: "html",
          //   position: "right",
          //   value: '<a href="https://ui.rilldata.com" class="navbar-cloud-btn" target="_blank" rel="noopener">Rill Cloud</a>',
          // },
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
    // @ts-ignore
    llmsTxtPlugin,
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
            to: '/get-started/install',
          },
          {
            from: '/home/example-repository',
            to: '/',
          },
          {
            from: '/develop/import-data',
            to: '/connect/'
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
            to: '/build/metrics-view/security'
          },
          {
            from: '/deploy/credentials/',
            to: '/connect/credentials/'
          },
          {
            from: '/build/credentials',
            to: '/connect/credentials/'
          },
          {
            from: '/deploy/credentials/s3',
            to: '/connect/data-source/s3'
          },
          {
            from: '/deploy/credentials/gcs',
            to: '/connect/data-source/gcs'
          },
          {
            from: '/deploy/credentials/azure',
            to: '/connect/data-source/azure'
          },
          {
            from: '/deploy/credentials/athena',
            to: '/connect/data-source/athena'
          },
          {
            from: '/deploy/credentials/bigquery',
            to: '/connect/data-source/bigquery'
          },
          {
            from: '/deploy/credentials/snowflake',
            to: '/connect/data-source/snowflake'
          },
          {
            from: '/deploy/credentials/postgres',
            to: '/connect/data-source/postgres'
          },
          {
            from: '/deploy/credentials/salesforce',
            to: '/connect/data-source/salesforce'
          },
          {
            from: '/deploy/credentials/motherduck',
            to: '/connect/data-source/duckdb'
          },
          {
            from: '/deploy/source-refresh',
            to: '/build/models/data-refresh'
          },
          {
            from: '/reference/templating',
            to: '/connect/templating'
          },
          {
            from: '/example-projects',
            to: '/#examples'
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
          // OLAP Engine redirects
          {
            from: '/reference/olap-engines/',
            to: '/connect/olap/'
          },
          {
            from: '/reference/olap-engines/duckdb',
            to: '/connect/olap/duckdb'
          },
          {
            from: '/reference/olap-engines/clickhouse',
            to: '/connect/olap/clickhouse'
          },
          {
            from: '/reference/olap-engines/pinot',
            to: '/connect/olap/pinot'
          },
          {
            from: '/reference/olap-engines/druid',
            to: '/connect/olap/druid'
          },
          {
            from: '/reference/olap-engines/multiple-olap',
            to: '/connect/olap/multiple-olap'
          },
          // Connector redirects
          {
            from: '/reference/connectors/',
            to: '/connect/'
          },
          {
            from: '/reference/connectors/gcs',
            to: '/connect/data-source/gcs'
          },
          {
            from: '/reference/connectors/azure',
            to: '/connect/data-source/azure'
          },
          {
            from: '/reference/connectors/s3',
            to: '/connect/data-source/s3'
          },
          {
            from: '/reference/connectors/snowflake',
            to: '/connect/data-source/snowflake'
          },
          {
            from: '/reference/connectors/bigquery',
            to: '/connect/data-source/bigquery'
          },
          {
            from: '/reference/connectors/redshift',
            to: '/connect/data-source/redshift'
          },
          {
            from: '/reference/connectors/postgres',
            to: '/connect/data-source/postgres'
          },
          {
            from: '/reference/connectors/athena',
            to: '/connect/data-source/athena'
          },
          {
            from: '/reference/connectors/mysql',
            to: '/connect/data-source/mysql'
          },
          {
            from: '/reference/connectors/sqlite',
            to: '/connect/data-source/sqlite'
          },
          {
            from: '/reference/connectors/salesforce',
            to: '/connect/data-source/salesforce'
          },
          {
            from: '/reference/connectors/sheets',
            to: '/connect/data-source/googlesheets'
          },
          {
            from: '/reference/connectors/slack',
            to: '/connect/data-source/slack'
          },
          {
            from: '/reference/connectors/local-file',
            to: '/connect/data-source/local-file'
          },
          {
            from: '/reference/connectors/https',
            to: '/connect/data-source/https'
          },
          // ADvand Model Redirects
          {
            from: '/reference/project-files/advanced-models',
            to: '/reference/project-files/models'
          },
          {
            from: '/deploy/templating',
            to: '/connect/templating'
          },
          {
            from: '/manage/account-management/billing',
            to: '/other/plans'
          },
          {
            from: '/manage/granting/azure-storage-container',
            to: '/other/granting/azure-storage-container'
          },
          {
            from: '/manage/granting/gcs-bucket',
            to: '/other/granting/gcs-bucket'
          },
          {
            from: '/manage/granting/google-bigquery',
            to: '/other/granting/google-bigquery'
          },
          {
            from: '/manage/granting/aws-s3-bucket',
            to: '/other/granting/aws-s3-bucket'
          },
          {
            from: '/manage/granting/',
            to: '/other/granting/'
          },
          {
            from: '/home/FAQ',
            to: '/other/FAQ'
          },
          {
            from: '/concepts/developerVsCloud',
            to: '/deploy/cloud-vs-developer'
          },
          {
            from: '/home/concepts/developerVsCloud',
            to: '/deploy/cloud-vs-developer'
          },
          {
            from: '/concepts/OLAP',
            to: '/connect/olap#what-is-olap'
          },
          {
            from: '/home/concepts/OLAP',
            to: '/connect/olap#what-is-olap'
          },
          {
            from: '/concepts/architecture',
            to: '/get-started/concepts/architecture'
          },
          {
            from: '/home/concepts/architecture',
            to: '/get-started/concepts/architecture'
          },
          {
            from: '/concepts/operational',
            to: '/get-started/concepts/operational'
          },
          {
            from: '/home/concepts/operational',
            to: '/get-started/concepts/operational'
          },
          {
            from: '/concepts/metrics-layer',
            to: '/build/metrics-view'
          },
          {
            from: '/concepts/bi-as-code',
            to: '/get-started/concepts/bi-as-code'
          },
          {
            from: '/home/concepts/bi-as-code',
            to: '/get-started/concepts/bi-as-code'
          },
          {
            from: '/build/advanced-models/',
            to: '/build/models/'
          },
          {
            from: '/build/advanced-models/incremental-models',
            to: '/build/models/incremental-models'
          },
          {
            from: '/build/advanced-models/partitions',
            to: '/build/models/partitioned-models'
          },
          {
            from: '/build/advanced-models/staging',
            to: '/build/models/staging-models'
          },
          {
            from: '/home/concepts/metrics-layer',
            to: '/build/metrics-view'
          },
          {
            from: '/integrate/custom-apis',
            to: '/build/custom-apis'
          },
          {
            from: '/deploy/performance',
            to: '/guides/performance'
          },
          {
            from: '/home/install',
            to: '/get-started/install'
          },
          {
            from: '/home/get-started',
            to: '/get-started/quickstart'
          },
          {
            from: '/deploy/deploy-dashboard/deploy-from-cli',
            to: '/guides/deploy-from-cli'
          },
          {
            from: '/deploy/deploy-dashboard/github-101',
            to: '/guides/github-101'
          },
          // {
          //   from: '/old-page',
          //   to: '/new-page',
          // }
        ],
      },
    ],
  ],

  // Configure Mermaid for diagrams
  themes: ['@docusaurus/theme-mermaid'],
  markdown: {
    mermaid: true,
  },
  stylesheets: [
    "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.0/css/all.min.css"
  ],
};

module.exports = config;
