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
            label: "Developer Docs",
            position: "left",
            className: "navbar-docs-link",
            activeBaseRegex: "^(?!/(reference|api|contact|notes|user-guide)).*", // Keep Docs active for all doc pages

          },
          {
            to: "/user-guide",
            label: "User Guide",
            position: "left",
            className: "navbar-user-guide-link",
            activeBaseRegex: "^/user-guide.*", // Keep Docs active for all doc pages
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
            to: '/developer/get-started/install',
          },
          {
            from: '/home/example-repository',
            to: '/',
          },
          {
            from: '/develop/import-data',
            to: '/developer/build/connectors/'
          },
          {
            from: '/develop/sql-models',
            to: '/developer/build/models'
          },
          {
            from: '/develop/metrics-dashboard',
            to: '/developer/build/dashboards'
          },
          {
            from: '/develop/security',
            to: '/developer/build/metrics-view/security'
          },
          {
            from: '/manage/security',
            to: '/developer/build/metrics-view/security'
          },
          {
            from: '/deploy/credentials/',
            to: '/developer/build/connectors/credentials/'
          },
          {
            from: '/build/credentials',
            to: '/developer/build/connectors/credentials/'
          },
          {
            from: '/deploy/credentials/s3',
            to: '/developer/build/connectors/data-source/s3'
          },
          {
            from: '/deploy/credentials/gcs',
            to: '/developer/build/connectors/data-source/gcs'
          },
          {
            from: '/deploy/credentials/azure',
            to: '/developer/build/connectors/data-source/azure'
          },
          {
            from: '/deploy/credentials/athena',
            to: '/developer/build/connectors/data-source/athena'
          },
          {
            from: '/deploy/credentials/bigquery',
            to: '/developer/build/connectors/data-source/bigquery'
          },
          {
            from: '/deploy/credentials/snowflake',
            to: '/developer/build/connectors/data-source/snowflake'
          },
          {
            from: '/deploy/credentials/postgres',
            to: '/developer/build/connectors/data-source/postgres'
          },
          {
            from: '/deploy/credentials/salesforce',
            to: '/developer/build/connectors/data-source/salesforce'
          },
          {
            from: '/deploy/credentials/motherduck',
            to: '/developer/build/connectors/olap/motherduck'
          },
          {
            from: '/deploy/source-refresh',
            to: '/developer/build/models/data-refresh'
          },
          {
            from: '/reference/templating',
            to: '/developer/build/connectors/templating'
          },
          {
            from: '/example-projects',
            to: '/#examples'
          },
          {
            from: '/integration/embedding',
            to: '/developer/integrate/embedding'
          },
          {
            from: '/share/user-management',
            to: '/user-guide/manage/user-management'
          },
          {
            from: '/share/roles-permissions',
            to: '/user-guide/manage/roles-permissions'
          },
          {
            from: '/share/scheduled-reports',
            to: '/user-guide/explore/exports'
          },
          // OLAP Engine redirects
          {
            from: '/reference/olap-engines/',
            to: '/developer/build/connectors/olap/'
          },
          {
            from: '/reference/olap-engines/duckdb',
            to: '/developer/build/connectors/olap/duckdb'
          },
          {
            from: '/reference/olap-engines/clickhouse',
            to: '/developer/build/connectors/olap/clickhouse'
          },
          {
            from: '/reference/olap-engines/pinot',
            to: '/developer/build/connectors/olap/pinot'
          },
          {
            from: '/reference/olap-engines/druid',
            to: '/developer/build/connectors/olap/druid'
          },
          {
            from: '/reference/olap-engines/multiple-olap',
            to: '/developer/build/connectors/olap/multiple-olap'
          },
          // Connector redirects
          {
            from: '/reference/connectors/',
            to: '/developer/build/connectors/'
          },
          {
            from: '/reference/connectors/gcs',
            to: '/developer/build/connectors/data-source/gcs'
          },
          {
            from: '/reference/connectors/azure',
            to: '/developer/build/connectors/data-source/azure'
          },
          {
            from: '/reference/connectors/s3',
            to: '/developer/build/connectors/data-source/s3'
          },
          {
            from: '/reference/connectors/snowflake',
            to: '/developer/build/connectors/data-source/snowflake'
          },
          {
            from: '/reference/connectors/bigquery',
            to: '/developer/build/connectors/data-source/bigquery'
          },
          {
            from: '/reference/connectors/redshift',
            to: '/developer/build/connectors/data-source/redshift'
          },
          {
            from: '/reference/connectors/postgres',
            to: '/developer/build/connectors/data-source/postgres'
          },
          {
            from: '/reference/connectors/athena',
            to: '/developer/build/connectors/data-source/athena'
          },
          {
            from: '/reference/connectors/mysql',
            to: '/developer/build/connectors/data-source/mysql'
          },
          {
            from: '/reference/connectors/sqlite',
            to: '/developer/build/connectors/data-source/sqlite'
          },
          {
            from: '/reference/connectors/salesforce',
            to: '/developer/build/connectors/data-source/salesforce'
          },
          {
            from: '/reference/connectors/sheets',
            to: '/developer/build/connectors/data-source/googlesheets'
          },
          {
            from: '/reference/connectors/slack',
            to: '/developer/build/connectors/data-source/slack'
          },
          {
            from: '/reference/connectors/local-file',
            to: '/developer/build/connectors/data-source/local-file'
          },
          {
            from: '/reference/connectors/https',
            to: '/developer/build/connectors/data-source/https'
          },
          // Advanced Model Redirects
          {
            from: '/reference/project-files/advanced-models',
            to: '/reference/project-files/models'
          },
          {
            from: '/deploy/templating',
            to: '/developer/build/connectors/templating'
          },
          {
            from: '/manage/account-management/billing',
            to: '/developer/other/plans'
          },
          {
            from: '/manage/granting/azure-storage-container',
            to: '/developer/other/granting/azure-storage-container'
          },
          {
            from: '/manage/granting/gcs-bucket',
            to: '/developer/other/granting/gcs-bucket'
          },
          {
            from: '/manage/granting/google-bigquery',
            to: '/developer/other/granting/google-bigquery'
          },
          {
            from: '/manage/granting/aws-s3-bucket',
            to: '/developer/other/granting/aws-s3-bucket'
          },
          {
            from: '/manage/granting/',
            to: '/developer/other/granting/'
          },
          {
            from: '/home/FAQ',
            to: '/developer/other/FAQ'
          },
          {
            from: '/concepts/developerVsCloud',
            to: '/developer/deploy/cloud-vs-developer'
          },
          {
            from: '/home/concepts/developerVsCloud',
            to: '/developer/deploy/cloud-vs-developer'
          },
          {
            from: '/get-started/concepts/cloud-vs-developer',
            to: '/developer/deploy/cloud-vs-developer'
          },
          {
            from: '/concepts/OLAP',
            to: '/developer/build/connectors/olap#what-is-olap'
          },
          {
            from: '/home/concepts/OLAP',
            to: '/developer/build/connectors/olap#what-is-olap'
          },
          {
            from: '/concepts/architecture',
            to: '/developer/get-started/why-rill#architecture'
          },
          {
            from: '/home/concepts/architecture',
            to: '/developer/get-started/why-rill#architecture'
          },
          {
            from: '/get-started/concepts/architecture',
            to: '/developer/get-started/why-rill#architecture'
          },
          {
            from: '/concepts/operational',
            to: '/developer/get-started/why-rill#operational-vs-traditional-bi'
          },
          {
            from: '/home/concepts/operational',
            to: '/developer/get-started/why-rill#operational-vs-traditional-bi'
          },
          {
            from: '/get-started/concepts/operational',
            to: '/developer/get-started/why-rill#operational-vs-traditional-bi'
          },
          {
            from: '/concepts/metrics-layer',
            to: '/developer/build/metrics-view'
          },
          {
            from: '/concepts/bi-as-code',
            to: '/developer/get-started/why-rill#bi-as-code'
          },
          {
            from: '/home/concepts/bi-as-code',
            to: '/developer/get-started/why-rill#bi-as-code'
          },
          {
            from: '/get-started/concepts/bi-as-code',
            to: '/developer/get-started/why-rill#bi-as-code'
          },
          {
            from: '/build/advanced-models/',
            to: '/developer/build/models/'
          },
          {
            from: '/build/advanced-models/incremental-models',
            to: '/developer/build/models/incremental-models'
          },
          {
            from: '/build/advanced-models/partitions',
            to: '/developer/build/models/partitioned-models'
          },
          {
            from: '/build/advanced-models/staging',
            to: '/developer/build/models/staging-models'
          },
          {
            from: '/home/concepts/metrics-layer',
            to: '/developer/build/metrics-view'
          },
          {
            from: '/integrate/custom-apis',
            to: '/developer/build/custom-apis'
          },
          {
            from: '/integrate/custom-apis/metrics-sql-api',
            to: '/developer/build/custom-apis'
          },
          {
            from: '/integrate/custom-apis/sql-api',
            to: '/developer/build/custom-apis'
          },
          {
            from: '/explore/filters/filters',
            to: '/user-guide/explore/filters'
          },
          {
            from: '/explore/filters/time-series',
            to: '/user-guide/explore/time-series'
          },
          {
            from: '/build/metrics-view/advanced-expressions/case-statements',
            to: '/developer/build/metrics-view/measures/case-statements'
          },
          {
            from: '/build/metrics-view/advanced-expressions/fixed-metrics',
            to: '/developer/build/metrics-view/measures/fixed-measures'
          },
          {
            from: '/build/metrics-view/advanced-expressions/metric-formatting',
            to: '/developer/build/metrics-view/measures/measures-formatting'
          },
          {
            from: '/build/metrics-view/advanced-expressions/quantiles',
            to: '/developer/build/metrics-view/measures/quantiles'
          },
          {
            from: '/build/metrics-view/advanced-expressions/referencing',
            to: '/developer/build/metrics-view/measures/referencing'
          },
          {
            from: '/build/metrics-view/advanced-expressions/unnesting',
            to: '/developer/build/metrics-view/dimensions/unnesting'
          },
          {
            from: '/build/metrics-view/advanced-expressions/windows',
            to: '/developer/build/metrics-view/measures/windows'
          },
          {
            from: '/build/metrics-view/advanced-expressions/advanced-expressions',
            to: '/developer/build/metrics-view/measures'
          },
          {
            from: '/build/metrics-view/customize',
            to: '/developer/build/metrics-view'
          },
          {
            from: '/deploy/performance',
            to: '/developer/guides/performance'
          },
          {
            from: '/home/install',
            to: '/developer/get-started/install'
          },
          {
            from: '/home/get-started',
            to: '/developer/get-started/quickstart'
          },
          {
            from: '/build/canvas/canvas',
            to: '/developer/build/dashboards/canvas',
          },
          {
            from: '/build/canvas/customization',
            to: '/developer/build/dashboards/customization',
          },
          {
            from: '/build/canvas',
            to: '/developer/build/dashboards/canvas',
          },
          // Redirect old /connect/ paths to new /build/connectors/ paths
          {
            from: '/connect',
            to: '/developer/build/connectors',
          },
          // Redirect /build/connect/ to /build/connectors/ for backward compatibility
          {
            from: '/build/connect',
            to: '/developer/build/connectors',
          },
          {
            from: '/build/connect/credentials',
            to: '/developer/build/connectors/credentials',
          },
          {
            from: '/build/connect/templating',
            to: '/developer/build/connectors/templating',
          },
          {
            from: '/build/connect/olap',
            to: '/developer/build/connectors/olap',
          },
          {
            from: '/build/connect/olap/duckdb',
            to: '/developer/build/connectors/olap/duckdb',
          },
          {
            from: '/build/connect/olap/clickhouse',
            to: '/developer/build/connectors/olap/clickhouse',
          },
          {
            from: '/build/connect/olap/druid',
            to: '/developer/build/connectors/olap/druid',
          },
          {
            from: '/build/connect/olap/pinot',
            to: '/developer/build/connectors/olap/pinot',
          },
          {
            from: '/build/connect/olap/motherduck',
            to: '/developer/build/connectors/olap/motherduck',
          },
          {
            from: '/build/connect/olap/multiple-olap',
            to: '/developer/build/connectors/olap/multiple-olap',
          },
          {
            from: '/build/connect/data-source',
            to: '/developer/build/connectors/data-source',
          },
          {
            from: '/build/connect/data-source/s3',
            to: '/developer/build/connectors/data-source/s3',
          },
          {
            from: '/build/connect/data-source/gcs',
            to: '/developer/build/connectors/data-source/gcs',
          },
          {
            from: '/build/connect/data-source/azure',
            to: '/developer/build/connectors/data-source/azure',
          },
          {
            from: '/build/connect/data-source/athena',
            to: '/developer/build/connectors/data-source/athena',
          },
          {
            from: '/build/connect/data-source/bigquery',
            to: '/developer/build/connectors/data-source/bigquery',
          },
          {
            from: '/build/connect/data-source/snowflake',
            to: '/developer/build/connectors/data-source/snowflake',
          },
          {
            from: '/build/connect/data-source/redshift',
            to: '/developer/build/connectors/data-source/redshift',
          },
          {
            from: '/build/connect/data-source/postgres',
            to: '/developer/build/connectors/data-source/postgres',
          },
          {
            from: '/build/connect/data-source/mysql',
            to: '/developer/build/connectors/data-source/mysql',
          },
          {
            from: '/build/connect/data-source/sqlite',
            to: '/developer/build/connectors/data-source/sqlite',
          },
          {
            from: '/build/connect/data-source/salesforce',
            to: '/developer/build/connectors/data-source/salesforce',
          },
          {
            from: '/build/connect/data-source/duckdb',
            to: '/developer/build/connectors/data-source/duckdb',
          },
          {
            from: '/build/connect/data-source/googlesheets',
            to: '/developer/build/connectors/data-source/googlesheets',
          },
          {
            from: '/build/connect/data-source/slack',
            to: '/developer/build/connectors/data-source/slack',
          },
          {
            from: '/build/connect/data-source/local-file',
            to: '/developer/build/connectors/data-source/local-file',
          },
          {
            from: '/build/connect/data-source/https',
            to: '/developer/build/connectors/data-source/https',
          },
          {
            from: '/build/connect/data-source/kafka',
            to: '/developer/build/connectors/data-source/kafka',
          },
          {
            from: '/build/connect/data-source/openai',
            to: '/developer/build/connectors/data-source/openai',
          },
          {
            from: '/connect/credentials',
            to: '/developer/build/connectors/credentials',
          },
          {
            from: '/connect/templating',
            to: '/developer/build/connectors/templating',
          },
          {
            from: '/connect/olap',
            to: '/developer/build/connectors/olap',
          },
          {
            from: '/connect/olap/duckdb',
            to: '/developer/build/connectors/olap/duckdb',
          },
          {
            from: '/connect/olap/clickhouse',
            to: '/developer/build/connectors/olap/clickhouse',
          },
          {
            from: '/connect/olap/druid',
            to: '/developer/build/connectors/olap/druid',
          },
          {
            from: '/connect/olap/pinot',
            to: '/developer/build/connectors/olap/pinot',
          },
          {
            from: '/connect/olap/motherduck',
            to: '/developer/build/connectors/olap/motherduck',
          },
          {
            from: '/connect/olap/multiple-olap',
            to: '/developer/build/connectors/olap/multiple-olap',
          },
          {
            from: '/connect/data-source',
            to: '/developer/build/connectors/data-source',
          },
          {
            from: '/connect/data-source/s3',
            to: '/developer/build/connectors/data-source/s3',
          },
          {
            from: '/connect/data-source/gcs',
            to: '/developer/build/connectors/data-source/gcs',
          },
          {
            from: '/connect/data-source/azure',
            to: '/developer/build/connectors/data-source/azure',
          },
          {
            from: '/connect/data-source/athena',
            to: '/developer/build/connectors/data-source/athena',
          },
          {
            from: '/connect/data-source/bigquery',
            to: '/developer/build/connectors/data-source/bigquery',
          },
          {
            from: '/connect/data-source/snowflake',
            to: '/developer/build/connectors/data-source/snowflake',
          },
          {
            from: '/connect/data-source/redshift',
            to: '/developer/build/connectors/data-source/redshift',
          },
          {
            from: '/connect/data-source/postgres',
            to: '/developer/build/connectors/data-source/postgres',
          },
          {
            from: '/connect/data-source/mysql',
            to: '/developer/build/connectors/data-source/mysql',
          },
          {
            from: '/connect/data-source/sqlite',
            to: '/developer/build/connectors/data-source/sqlite',
          },
          {
            from: '/connect/data-source/salesforce',
            to: '/developer/build/connectors/data-source/salesforce',
          },
          {
            from: '/connect/data-source/duckdb',
            to: '/developer/build/connectors/data-source/duckdb',
          },
          {
            from: '/connect/data-source/googlesheets',
            to: '/developer/build/connectors/data-source/googlesheets',
          },
          {
            from: '/connect/data-source/slack',
            to: '/developer/build/connectors/data-source/slack',
          },
          {
            from: '/connect/data-source/local-file',
            to: '/developer/build/connectors/data-source/local-file',
          },
          {
            from: '/connect/data-source/https',
            to: '/developer/build/connectors/data-source/https',
          },
          {
            from: '/connect/data-source/kafka',
            to: '/developer/build/connectors/data-source/kafka',
          },
          {
            from: '/connect/data-source/openai',
            to: '/developer/build/connectors/data-source/openai',
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
