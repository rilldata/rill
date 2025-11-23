// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

/* eslint @typescript-eslint/no-var-requires: "off" */
const { themes } = require('prism-react-renderer');
const lightCodeTheme = themes.github;
const darkCodeTheme = themes.dracula;

const llmsTxtPlugin = require('./plugins/llms-txt-plugin');

const def = require("redocusaurus");
const path = require('path');
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
    [
      'redocusaurus',
      {
        config: path.join(__dirname, 'redocly.yaml'),
        specs: [
          {
            id: 'admin',
            spec: '../proto/gen/rill/admin/v1/public.openapi.yaml',
            route: '/api/admin/',
          },
        ],
        theme: {
          primaryColor: '#3524c7',
        },
      },
    ]
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

        // Navbar button text (before clicking)
        translations: {
          button: {
            buttonText: 'Search...',
            buttonAriaLabel: 'Open search',
          },
        },
        placeholder: "Looking for something?",
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
            to: "/",
            label: "Docs",
            position: "left",
            className: "navbar-docs-link",
            activeBaseRegex: "^(?!/(reference|api|contact|notes)).*", // Keep Docs active for all doc pages
          },
          {
            to: "/reference/project-files",
            label: "Reference",
            position: "left",
            className: "navbar-reference-link",
            activeBasePath: "/reference",
          },
          {
            to: "/api/admin/",
            label: "API",
            position: "left",
            className: "navbar-api-link",
            activeBasePath: "/api/admin",
          },
          {
            to: "/contact",
            label: "Contact Us",
            position: "left",
            className: "navbar-contact-link",
            activeBasePath: "/contact",
          },



          // Right side items
          {
            type: "html",
            position: "right",
            value: '<a href="https://github.com/rilldata/rill" class="navbar-icon-link" aria-label="GitHub" target="_blank" rel="noopener noreferrer">GitHub</a>',
          },
          {
            type: "html",
            position: "right",
            value: '<a href="https://www.rilldata.com/blog" class="navbar-icon-link" aria-label="Blog" target="_blank" rel="noopener noreferrer">Blog</a>',
          },

          {
            type: "search",
            position: "right"
          },
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
            to: '/build/connectors/'
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
            from: '/manage/security',
            to: '/build/metrics-view/security'
          },
          {
            from: '/deploy/credentials/',
            to: '/build/connectors/credentials/'
          },
          {
            from: '/build/credentials',
            to: '/build/connectors/credentials/'
          },
          {
            from: '/deploy/credentials/s3',
            to: '/build/connectors/data-source/s3'
          },
          {
            from: '/deploy/credentials/gcs',
            to: '/build/connectors/data-source/gcs'
          },
          {
            from: '/deploy/credentials/azure',
            to: '/build/connectors/data-source/azure'
          },
          {
            from: '/deploy/credentials/athena',
            to: '/build/connectors/data-source/athena'
          },
          {
            from: '/deploy/credentials/bigquery',
            to: '/build/connectors/data-source/bigquery'
          },
          {
            from: '/deploy/credentials/snowflake',
            to: '/build/connectors/data-source/snowflake'
          },
          {
            from: '/deploy/credentials/postgres',
            to: '/build/connectors/data-source/postgres'
          },
          {
            from: '/deploy/credentials/salesforce',
            to: '/build/connectors/data-source/salesforce'
          },
          {
            from: '/deploy/credentials/motherduck',
            to: '/build/connectors/olap/motherduck'
          },
          {
            from: '/deploy/source-refresh',
            to: '/build/models/data-refresh'
          },
          {
            from: '/reference/templating',
            to: '/build/connectors/templating'
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
            to: '/build/connectors/olap/'
          },
          {
            from: '/reference/olap-engines/duckdb',
            to: '/build/connectors/olap/duckdb'
          },
          {
            from: '/reference/olap-engines/clickhouse',
            to: '/build/connectors/olap/clickhouse'
          },
          {
            from: '/reference/olap-engines/pinot',
            to: '/build/connectors/olap/pinot'
          },
          {
            from: '/reference/olap-engines/druid',
            to: '/build/connectors/olap/druid'
          },
          {
            from: '/reference/olap-engines/multiple-olap',
            to: '/build/connectors/olap/multiple-olap'
          },
          // Connector redirects
          {
            from: '/reference/connectors/',
            to: '/build/connectors/'
          },
          {
            from: '/reference/connectors/gcs',
            to: '/build/connectors/data-source/gcs'
          },
          {
            from: '/reference/connectors/azure',
            to: '/build/connectors/data-source/azure'
          },
          {
            from: '/reference/connectors/s3',
            to: '/build/connectors/data-source/s3'
          },
          {
            from: '/reference/connectors/snowflake',
            to: '/build/connectors/data-source/snowflake'
          },
          {
            from: '/reference/connectors/bigquery',
            to: '/build/connectors/data-source/bigquery'
          },
          {
            from: '/reference/connectors/redshift',
            to: '/build/connectors/data-source/redshift'
          },
          {
            from: '/reference/connectors/postgres',
            to: '/build/connectors/data-source/postgres'
          },
          {
            from: '/reference/connectors/athena',
            to: '/build/connectors/data-source/athena'
          },
          {
            from: '/reference/connectors/mysql',
            to: '/build/connectors/data-source/mysql'
          },
          {
            from: '/reference/connectors/sqlite',
            to: '/build/connectors/data-source/sqlite'
          },
          {
            from: '/reference/connectors/salesforce',
            to: '/build/connectors/data-source/salesforce'
          },
          {
            from: '/reference/connectors/sheets',
            to: '/build/connectors/data-source/googlesheets'
          },
          {
            from: '/reference/connectors/slack',
            to: '/build/connectors/data-source/slack'
          },
          {
            from: '/reference/connectors/local-file',
            to: '/build/connectors/data-source/local-file'
          },
          {
            from: '/reference/connectors/https',
            to: '/build/connectors/data-source/https'
          },
          // ADvand Model Redirects
          {
            from: '/reference/project-files/advanced-models',
            to: '/reference/project-files/models'
          },
          {
            from: '/deploy/templating',
            to: '/build/connectors/templating'
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
            to: '/get-started/concepts/cloud-vs-developer'
          },
          {
            from: '/home/concepts/developerVsCloud',
            to: '/get-started/concepts/cloud-vs-developer'
          },
          {
            from: '/concepts/OLAP',
            to: '/build/connectors/olap#what-is-olap'
          },
          {
            from: '/home/concepts/OLAP',
            to: '/build/connectors/olap#what-is-olap'
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
            from: '/integrate/custom-apis/metrics-sql-api',
            to: '/build/custom-apis'
          },
          {
            from: '/integrate/custom-apis/sql-api',
            to: '/build/custom-apis'
          },
          {
            from: '/explore/filters/filters',
            to: '/explore/filters'
          },
          {
            from: '/explore/filters/time-series',
            to: '/explore/time-series'
          },
          {
            from: '/build/metrics-view/advanced-expressions/case-statements',
            to: '/build/metrics-view/measures/case-statements'
          },
          {
            from: '/build/metrics-view/advanced-expressions/fixed-metrics',
            to: '/build/metrics-view/measures/fixed-measures'
          },
          {
            from: '/build/metrics-view/advanced-expressions/metric-formatting',
            to: '/build/metrics-view/measures/measures-formatting'
          },
          {
            from: '/build/metrics-view/advanced-expressions/quantiles',
            to: '/build/metrics-view/measures/quantiles'
          },
          {
            from: '/build/metrics-view/advanced-expressions/referencing',
            to: '/build/metrics-view/measures/referencing'
          },
          {
            from: '/build/metrics-view/advanced-expressions/unnesting',
            to: '/build/metrics-view/dimensions/unnesting'
          },
          {
            from: '/build/metrics-view/advanced-expressions/windows',
            to: '/build/metrics-view/measures/windows'
          },
          {
            from: '/build/metrics-view/advanced-expressions/advanced-expressions',
            to: '/build/metrics-view/measures'
          },
          {
            from: '/build/metrics-view/customize',
            to: '/build/metrics-view'
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
            from: '/build/canvas/canvas',
            to: '/build/dashboards/canvas',
          },
          {
            from: '/build/canvas/customization',
            to: '/build/dashboards/customization',
          },
          {
            from: '/build/canvas',
            to: '/build/dashboards/canvas',
          },
          // Redirect old /connect/ paths to new /build/connectors/ paths
          {
            from: '/connect',
            to: '/build/connectors',
          },
          // Redirect /build/connect/ to /build/connectors/ for backward compatibility
          {
            from: '/build/connect',
            to: '/build/connectors',
          },
          {
            from: '/build/connect/credentials',
            to: '/build/connectors/credentials',
          },
          {
            from: '/build/connect/templating',
            to: '/build/connectors/templating',
          },
          {
            from: '/build/connect/olap',
            to: '/build/connectors/olap',
          },
          {
            from: '/build/connect/olap/duckdb',
            to: '/build/connectors/olap/duckdb',
          },
          {
            from: '/build/connect/olap/clickhouse',
            to: '/build/connectors/olap/clickhouse',
          },
          {
            from: '/build/connect/olap/druid',
            to: '/build/connectors/olap/druid',
          },
          {
            from: '/build/connect/olap/pinot',
            to: '/build/connectors/olap/pinot',
          },
          {
            from: '/build/connect/olap/motherduck',
            to: '/build/connectors/olap/motherduck',
          },
          {
            from: '/build/connect/olap/multiple-olap',
            to: '/build/connectors/olap/multiple-olap',
          },
          {
            from: '/build/connect/data-source',
            to: '/build/connectors/data-source',
          },
          {
            from: '/build/connect/data-source/s3',
            to: '/build/connectors/data-source/s3',
          },
          {
            from: '/build/connect/data-source/gcs',
            to: '/build/connectors/data-source/gcs',
          },
          {
            from: '/build/connect/data-source/azure',
            to: '/build/connectors/data-source/azure',
          },
          {
            from: '/build/connect/data-source/athena',
            to: '/build/connectors/data-source/athena',
          },
          {
            from: '/build/connect/data-source/bigquery',
            to: '/build/connectors/data-source/bigquery',
          },
          {
            from: '/build/connect/data-source/snowflake',
            to: '/build/connectors/data-source/snowflake',
          },
          {
            from: '/build/connect/data-source/redshift',
            to: '/build/connectors/data-source/redshift',
          },
          {
            from: '/build/connect/data-source/postgres',
            to: '/build/connectors/data-source/postgres',
          },
          {
            from: '/build/connect/data-source/mysql',
            to: '/build/connectors/data-source/mysql',
          },
          {
            from: '/build/connect/data-source/sqlite',
            to: '/build/connectors/data-source/sqlite',
          },
          {
            from: '/build/connect/data-source/salesforce',
            to: '/build/connectors/data-source/salesforce',
          },
          {
            from: '/build/connect/data-source/duckdb',
            to: '/build/connectors/data-source/duckdb',
          },
          {
            from: '/build/connect/data-source/googlesheets',
            to: '/build/connectors/data-source/googlesheets',
          },
          {
            from: '/build/connect/data-source/slack',
            to: '/build/connectors/data-source/slack',
          },
          {
            from: '/build/connect/data-source/local-file',
            to: '/build/connectors/data-source/local-file',
          },
          {
            from: '/build/connect/data-source/https',
            to: '/build/connectors/data-source/https',
          },
          {
            from: '/build/connect/data-source/kafka',
            to: '/build/connectors/data-source/kafka',
          },
          {
            from: '/build/connect/data-source/openai',
            to: '/build/connectors/data-source/openai',
          },
          {
            from: '/connect/credentials',
            to: '/build/connectors/credentials',
          },
          {
            from: '/connect/templating',
            to: '/build/connectors/templating',
          },
          {
            from: '/connect/olap',
            to: '/build/connectors/olap',
          },
          {
            from: '/connect/olap/duckdb',
            to: '/build/connectors/olap/duckdb',
          },
          {
            from: '/connect/olap/clickhouse',
            to: '/build/connectors/olap/clickhouse',
          },
          {
            from: '/connect/olap/druid',
            to: '/build/connectors/olap/druid',
          },
          {
            from: '/connect/olap/pinot',
            to: '/build/connectors/olap/pinot',
          },
          {
            from: '/connect/olap/motherduck',
            to: '/build/connectors/olap/motherduck',
          },
          {
            from: '/connect/olap/multiple-olap',
            to: '/build/connectors/olap/multiple-olap',
          },
          {
            from: '/connect/data-source',
            to: '/build/connectors/data-source',
          },
          {
            from: '/connect/data-source/s3',
            to: '/build/connectors/data-source/s3',
          },
          {
            from: '/connect/data-source/gcs',
            to: '/build/connectors/data-source/gcs',
          },
          {
            from: '/connect/data-source/azure',
            to: '/build/connectors/data-source/azure',
          },
          {
            from: '/connect/data-source/athena',
            to: '/build/connectors/data-source/athena',
          },
          {
            from: '/connect/data-source/bigquery',
            to: '/build/connectors/data-source/bigquery',
          },
          {
            from: '/connect/data-source/snowflake',
            to: '/build/connectors/data-source/snowflake',
          },
          {
            from: '/connect/data-source/redshift',
            to: '/build/connectors/data-source/redshift',
          },
          {
            from: '/connect/data-source/postgres',
            to: '/build/connectors/data-source/postgres',
          },
          {
            from: '/connect/data-source/mysql',
            to: '/build/connectors/data-source/mysql',
          },
          {
            from: '/connect/data-source/sqlite',
            to: '/build/connectors/data-source/sqlite',
          },
          {
            from: '/connect/data-source/salesforce',
            to: '/build/connectors/data-source/salesforce',
          },
          {
            from: '/connect/data-source/duckdb',
            to: '/build/connectors/data-source/duckdb',
          },
          {
            from: '/connect/data-source/googlesheets',
            to: '/build/connectors/data-source/googlesheets',
          },
          {
            from: '/connect/data-source/slack',
            to: '/build/connectors/data-source/slack',
          },
          {
            from: '/connect/data-source/local-file',
            to: '/build/connectors/data-source/local-file',
          },
          {
            from: '/connect/data-source/https',
            to: '/build/connectors/data-source/https',
          },
          {
            from: '/connect/data-source/kafka',
            to: '/build/connectors/data-source/kafka',
          },
          {
            from: '/connect/data-source/openai',
            to: '/build/connectors/data-source/openai',
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
    hooks: {
      onBrokenMarkdownLinks: "throw",
    },
  },
  stylesheets: [
    "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.0/css/all.min.css"
  ],

};

module.exports = config;
