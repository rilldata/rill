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
            id: 'public',
            spec: 'api/openapi.yaml',
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
            type: "docSidebar",
            sidebarId: "developersSidebar",
            position: "left",
            className: "navbar-docs-link",
            label: "Developer Docs",
            activeBaseRegex: "^(?!/(reference|api|contact|notes)).*", // Keep Docs active for all doc pages

          },
          {
            type: "docSidebar",
            sidebarId: "usersSidebar",
            position: "left",
            className: "navbar-docs-link",
            label: "User Guide",
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
            to: '/developers/get-started/install',
          },
          {
            from: '/home/example-repository',
            to: '/',
          },
          {
            from: '/develop/import-data',
            to: '/developers/build/connectors/'
          },
          {
            from: '/develop/sql-models',
            to: '/developers/build/models'
          },
          {
            from: '/develop/metrics-dashboard',
            to: '/developers/build/dashboards'
          },
          {
            from: '/develop/security',
            to: '/developers/build/metrics-view/security'
          },
          {
            from: '/manage/security',
            to: '/developers/build/metrics-view/security'
          },
          {
            from: '/deploy/credentials/',
            to: '/developers/build/connectors/credentials/'
          },
          {
            from: '/build/credentials',
            to: '/developers/build/connectors/credentials/'
          },
          {
            from: '/deploy/credentials/s3',
            to: '/developers/build/connectors/data-source/s3'
          },
          {
            from: '/deploy/credentials/gcs',
            to: '/developers/build/connectors/data-source/gcs'
          },
          {
            from: '/deploy/credentials/azure',
            to: '/developers/build/connectors/data-source/azure'
          },
          {
            from: '/deploy/credentials/athena',
            to: '/developers/build/connectors/data-source/athena'
          },
          {
            from: '/deploy/credentials/bigquery',
            to: '/developers/build/connectors/data-source/bigquery'
          },
          {
            from: '/deploy/credentials/snowflake',
            to: '/developers/build/connectors/data-source/snowflake'
          },
          {
            from: '/deploy/credentials/postgres',
            to: '/developers/build/connectors/data-source/postgres'
          },
          {
            from: '/deploy/credentials/salesforce',
            to: '/developers/build/connectors/data-source/salesforce'
          },
          {
            from: '/deploy/credentials/motherduck',
            to: '/developers/build/connectors/olap/motherduck'
          },
          {
            from: '/deploy/source-refresh',
            to: '/developers/build/models/data-refresh'
          },
          {
            from: '/reference/templating',
            to: '/developers/build/connectors/templating'
          },
          {
            from: '/example-projects',
            to: '/#examples'
          },
          {
            from: '/integration/embedding',
            to: '/developers/integrate/embedding'
          },
          {
            from: '/share/user-management',
            to: '/users/manage/user-management'
          },
          {
            from: '/share/roles-permissions',
            to: '/users/manage/roles-permissions'
          },
          {
            from: '/share/scheduled-reports',
            to: '/users/explore/exports'
          },
          // OLAP Engine redirects
          {
            from: '/reference/olap-engines/',
            to: '/developers/build/connectors/olap/'
          },
          {
            from: '/reference/olap-engines/duckdb',
            to: '/developers/build/connectors/olap/duckdb'
          },
          {
            from: '/reference/olap-engines/clickhouse',
            to: '/developers/build/connectors/olap/clickhouse'
          },
          {
            from: '/reference/olap-engines/pinot',
            to: '/developers/build/connectors/olap/pinot'
          },
          {
            from: '/reference/olap-engines/druid',
            to: '/developers/build/connectors/olap/druid'
          },
          {
            from: '/reference/olap-engines/multiple-olap',
            to: '/developers/build/connectors/olap/multiple-olap'
          },
          // Connector redirects
          {
            from: '/reference/connectors/',
            to: '/developers/build/connectors/'
          },
          {
            from: '/reference/connectors/gcs',
            to: '/developers/build/connectors/data-source/gcs'
          },
          {
            from: '/reference/connectors/azure',
            to: '/developers/build/connectors/data-source/azure'
          },
          {
            from: '/reference/connectors/s3',
            to: '/developers/build/connectors/data-source/s3'
          },
          {
            from: '/reference/connectors/snowflake',
            to: '/developers/build/connectors/data-source/snowflake'
          },
          {
            from: '/reference/connectors/bigquery',
            to: '/developers/build/connectors/data-source/bigquery'
          },
          {
            from: '/reference/connectors/redshift',
            to: '/developers/build/connectors/data-source/redshift'
          },
          {
            from: '/reference/connectors/postgres',
            to: '/developers/build/connectors/data-source/postgres'
          },
          {
            from: '/reference/connectors/athena',
            to: '/developers/build/connectors/data-source/athena'
          },
          {
            from: '/reference/connectors/mysql',
            to: '/developers/build/connectors/data-source/mysql'
          },
          {
            from: '/reference/connectors/sqlite',
            to: '/developers/build/connectors/data-source/sqlite'
          },
          {
            from: '/reference/connectors/salesforce',
            to: '/developers/build/connectors/data-source/salesforce'
          },
          {
            from: '/reference/connectors/sheets',
            to: '/developers/build/connectors/data-source/googlesheets'
          },
          {
            from: '/reference/connectors/slack',
            to: '/developers/build/connectors/data-source/slack'
          },
          {
            from: '/reference/connectors/local-file',
            to: '/developers/build/connectors/data-source/local-file'
          },
          {
            from: '/reference/connectors/https',
            to: '/developers/build/connectors/data-source/https'
          },
          // ADvand Model Redirects
          {
            from: '/reference/project-files/advanced-models',
            to: '/reference/project-files/models'
          },
          {
            from: '/deploy/templating',
            to: '/developers/build/connectors/templating'
          },
          {
            from: '/manage/account-management/billing',
            to: '/developers/other/plans'
          },
          {
            from: '/manage/granting/azure-storage-container',
            to: '/developers/other/granting/azure-storage-container'
          },
          {
            from: '/manage/granting/gcs-bucket',
            to: '/developers/other/granting/gcs-bucket'
          },
          {
            from: '/manage/granting/google-bigquery',
            to: '/developers/other/granting/google-bigquery'
          },
          {
            from: '/manage/granting/aws-s3-bucket',
            to: '/developers/other/granting/aws-s3-bucket'
          },
          {
            from: '/manage/granting/',
            to: '/developers/other/granting/'
          },
          {
            from: '/home/FAQ',
            to: '/developers/other/FAQ'
          },
          {
            from: '/concepts/developerVsCloud',
            to: '/developers/get-started/concepts/cloud-vs-developer'
          },
          {
            from: '/home/concepts/developerVsCloud',
            to: '/developers/get-started/concepts/cloud-vs-developer'
          },
          {
            from: '/concepts/OLAP',
            to: '/developers/build/connectors/olap#what-is-olap'
          },
          {
            from: '/home/concepts/OLAP',
            to: '/developers/build/connectors/olap#what-is-olap'
          },
          {
            from: '/concepts/architecture',
            to: '/developers/get-started/concepts/architecture'
          },
          {
            from: '/home/concepts/architecture',
            to: '/developers/get-started/concepts/architecture'
          },
          {
            from: '/concepts/operational',
            to: '/developers/get-started/concepts/operational'
          },
          {
            from: '/home/concepts/operational',
            to: '/developers/get-started/concepts/operational'
          },
          {
            from: '/concepts/metrics-layer',
            to: '/developers/build/metrics-view'
          },
          {
            from: '/concepts/bi-as-code',
            to: '/developers/get-started/concepts/bi-as-code'
          },
          {
            from: '/home/concepts/bi-as-code',
            to: '/developers/get-started/concepts/bi-as-code'
          },
          {
            from: '/build/advanced-models/',
            to: '/developers/build/models/'
          },
          {
            from: '/build/advanced-models/incremental-models',
            to: '/developers/build/models/incremental-models'
          },
          {
            from: '/build/advanced-models/partitions',
            to: '/developers/build/models/partitioned-models'
          },
          {
            from: '/build/advanced-models/staging',
            to: '/developers/build/models/staging-models'
          },
          {
            from: '/home/concepts/metrics-layer',
            to: '/developers/build/metrics-view'
          },
          {
            from: '/integrate/custom-apis',
            to: '/developers/build/custom-apis'
          },
          {
            from: '/integrate/custom-apis/metrics-sql-api',
            to: '/developers/build/custom-apis'
          },
          {
            from: '/integrate/custom-apis/sql-api',
            to: '/developers/build/custom-apis'
          },
          {
            from: '/explore/filters/filters',
            to: '/users/explore/filters'
          },
          {
            from: '/explore/filters/time-series',
            to: '/users/explore/time-series'
          },
          {
            from: '/build/metrics-view/advanced-expressions/case-statements',
            to: '/developers/build/metrics-view/measures/case-statements'
          },
          {
            from: '/build/metrics-view/advanced-expressions/fixed-metrics',
            to: '/developers/build/metrics-view/measures/fixed-measures'
          },
          {
            from: '/build/metrics-view/advanced-expressions/metric-formatting',
            to: '/developers/build/metrics-view/measures/measures-formatting'
          },
          {
            from: '/build/metrics-view/advanced-expressions/quantiles',
            to: '/developers/build/metrics-view/measures/quantiles'
          },
          {
            from: '/build/metrics-view/advanced-expressions/referencing',
            to: '/developers/build/metrics-view/measures/referencing'
          },
          {
            from: '/build/metrics-view/advanced-expressions/unnesting',
            to: '/developers/build/metrics-view/dimensions/unnesting'
          },
          {
            from: '/build/metrics-view/advanced-expressions/windows',
            to: '/developers/build/metrics-view/measures/windows'
          },
          {
            from: '/build/metrics-view/advanced-expressions/advanced-expressions',
            to: '/developers/build/metrics-view/measures'
          },
          {
            from: '/build/metrics-view/customize',
            to: '/developers/build/metrics-view'
          },
          {
            from: '/deploy/performance',
            to: '/developers/guides/performance'
          },
          {
            from: '/home/install',
            to: '/developers/get-started/install'
          },
          {
            from: '/home/get-started',
            to: '/developers/get-started/quickstart'
          },
          {
            from: '/build/canvas/canvas',
            to: '/developers/build/dashboards/canvas',
          },
          {
            from: '/build/canvas/customization',
            to: '/developers/build/dashboards/customization',
          },
          {
            from: '/build/canvas',
            to: '/developers/build/dashboards/canvas',
          },
          // Redirect old /connect/ paths to new /build/connectors/ paths
          {
            from: '/connect',
            to: '/developers/build/connectors',
          },
          // Redirect /build/connect/ to /build/connectors/ for backward compatibility
          {
            from: '/build/connect',
            to: '/developers/build/connectors',
          },
          {
            from: '/build/connect/credentials',
            to: '/developers/build/connectors/credentials',
          },
          {
            from: '/build/connect/templating',
            to: '/developers/build/connectors/templating',
          },
          {
            from: '/build/connect/olap',
            to: '/developers/build/connectors/olap',
          },
          {
            from: '/build/connect/olap/duckdb',
            to: '/developers/build/connectors/olap/duckdb',
          },
          {
            from: '/build/connect/olap/clickhouse',
            to: '/developers/build/connectors/olap/clickhouse',
          },
          {
            from: '/build/connect/olap/druid',
            to: '/developers/build/connectors/olap/druid',
          },
          {
            from: '/build/connect/olap/pinot',
            to: '/developers/build/connectors/olap/pinot',
          },
          {
            from: '/build/connect/olap/motherduck',
            to: '/developers/build/connectors/olap/motherduck',
          },
          {
            from: '/build/connect/olap/multiple-olap',
            to: '/developers/build/connectors/olap/multiple-olap',
          },
          {
            from: '/build/connect/data-source',
            to: '/developers/build/connectors/data-source',
          },
          {
            from: '/build/connect/data-source/s3',
            to: '/developers/build/connectors/data-source/s3',
          },
          {
            from: '/build/connect/data-source/gcs',
            to: '/developers/build/connectors/data-source/gcs',
          },
          {
            from: '/build/connect/data-source/azure',
            to: '/developers/build/connectors/data-source/azure',
          },
          {
            from: '/build/connect/data-source/athena',
            to: '/developers/build/connectors/data-source/athena',
          },
          {
            from: '/build/connect/data-source/bigquery',
            to: '/developers/build/connectors/data-source/bigquery',
          },
          {
            from: '/build/connect/data-source/snowflake',
            to: '/developers/build/connectors/data-source/snowflake',
          },
          {
            from: '/build/connect/data-source/redshift',
            to: '/developers/build/connectors/data-source/redshift',
          },
          {
            from: '/build/connect/data-source/postgres',
            to: '/developers/build/connectors/data-source/postgres',
          },
          {
            from: '/build/connect/data-source/mysql',
            to: '/developers/build/connectors/data-source/mysql',
          },
          {
            from: '/build/connect/data-source/sqlite',
            to: '/developers/build/connectors/data-source/sqlite',
          },
          {
            from: '/build/connect/data-source/salesforce',
            to: '/developers/build/connectors/data-source/salesforce',
          },
          {
            from: '/build/connect/data-source/duckdb',
            to: '/developers/build/connectors/data-source/duckdb',
          },
          {
            from: '/build/connect/data-source/googlesheets',
            to: '/developers/build/connectors/data-source/googlesheets',
          },
          {
            from: '/build/connect/data-source/slack',
            to: '/developers/build/connectors/data-source/slack',
          },
          {
            from: '/build/connect/data-source/local-file',
            to: '/developers/build/connectors/data-source/local-file',
          },
          {
            from: '/build/connect/data-source/https',
            to: '/developers/build/connectors/data-source/https',
          },
          {
            from: '/build/connect/data-source/kafka',
            to: '/developers/build/connectors/data-source/kafka',
          },
          {
            from: '/build/connect/data-source/openai',
            to: '/developers/build/connectors/data-source/openai',
          },
          {
            from: '/connect/credentials',
            to: '/developers/build/connectors/credentials',
          },
          {
            from: '/connect/templating',
            to: '/developers/build/connectors/templating',
          },
          {
            from: '/connect/olap',
            to: '/developers/build/connectors/olap',
          },
          {
            from: '/connect/olap/duckdb',
            to: '/developers/build/connectors/olap/duckdb',
          },
          {
            from: '/connect/olap/clickhouse',
            to: '/developers/build/connectors/olap/clickhouse',
          },
          {
            from: '/connect/olap/druid',
            to: '/developers/build/connectors/olap/druid',
          },
          {
            from: '/connect/olap/pinot',
            to: '/developers/build/connectors/olap/pinot',
          },
          {
            from: '/connect/olap/motherduck',
            to: '/developers/build/connectors/olap/motherduck',
          },
          {
            from: '/connect/olap/multiple-olap',
            to: '/developers/build/connectors/olap/multiple-olap',
          },
          {
            from: '/connect/data-source',
            to: '/developers/build/connectors/data-source',
          },
          {
            from: '/connect/data-source/s3',
            to: '/developers/build/connectors/data-source/s3',
          },
          {
            from: '/connect/data-source/gcs',
            to: '/developers/build/connectors/data-source/gcs',
          },
          {
            from: '/connect/data-source/azure',
            to: '/developers/build/connectors/data-source/azure',
          },
          {
            from: '/connect/data-source/athena',
            to: '/developers/build/connectors/data-source/athena',
          },
          {
            from: '/connect/data-source/bigquery',
            to: '/developers/build/connectors/data-source/bigquery',
          },
          {
            from: '/connect/data-source/snowflake',
            to: '/developers/build/connectors/data-source/snowflake',
          },
          {
            from: '/connect/data-source/redshift',
            to: '/developers/build/connectors/data-source/redshift',
          },
          {
            from: '/connect/data-source/postgres',
            to: '/developers/build/connectors/data-source/postgres',
          },
          {
            from: '/connect/data-source/mysql',
            to: '/developers/build/connectors/data-source/mysql',
          },
          {
            from: '/connect/data-source/sqlite',
            to: '/developers/build/connectors/data-source/sqlite',
          },
          {
            from: '/connect/data-source/salesforce',
            to: '/developers/build/connectors/data-source/salesforce',
          },
          {
            from: '/connect/data-source/duckdb',
            to: '/developers/build/connectors/data-source/duckdb',
          },
          {
            from: '/connect/data-source/googlesheets',
            to: '/developers/build/connectors/data-source/googlesheets',
          },
          {
            from: '/connect/data-source/slack',
            to: '/developers/build/connectors/data-source/slack',
          },
          {
            from: '/connect/data-source/local-file',
            to: '/developers/build/connectors/data-source/local-file',
          },
          {
            from: '/connect/data-source/https',
            to: '/developers/build/connectors/data-source/https',
          },
          {
            from: '/connect/data-source/kafka',
            to: '/developers/build/connectors/data-source/kafka',
          },
          {
            from: '/connect/data-source/openai',
            to: '/developers/build/connectors/data-source/openai',
          },

          // Build section moved to developers/build
          {
            from: '/build',
            to: '/developers/build',
          },
          {
            from: '/build/getting-started',
            to: '/developers/build/rill-project',
          },
          {
            from: '/build/project-configuration',
            to: '/developers/build/project-configuration',
          },

          // Deploy section moved to developers/deploy
          {
            from: '/deploy',
            to: '/developers/deploy',
          },
          {
            from: '/deploy/deploy-credentials',
            to: '/developers/deploy/deploy-credentials',
          },
          {
            from: '/deploy/project-errors',
            to: '/developers/deploy/project-errors',
          },

          // Get started section
          {
            from: '/get-started',
            to: '/',
          },
          {
            from: '/get-started/get-started',
            to: '/',
          },
          {
            from: '/get-started/install',
            to: '/developers/get-started/install',
          },

          // Home to get-started
          {
            from: '/home',
            to: '/',
          },
          {
            from: '/home/home',
            to: '/',
          },

          // Guides section
          {
            from: '/guides',
            to: '/developers/guides',
          },
          {
            from: '/guides/clone-a-project',
            to: '/developers/guides/clone-a-project',
          },
          {
            from: '/guides/performance',
            to: '/developers/guides/performance',
          },

          // Integrate section
          {
            from: '/integrate',
            to: '/developers/integrate',
          },
          {
            from: '/integrate/custom-api',
            to: '/developers/integrate/custom-api',
          },
          {
            from: '/integrate/embed-api',
            to: '/developers/integrate/embed-iframe-api',
          },
          {
            from: '/integrate/embedding',
            to: '/developers/integrate/embedding',
          },

          // Other section
          {
            from: '/other',
            to: '/developers/other/FAQ',
          },
          {
            from: '/other/FAQ',
            to: '/developers/other/FAQ',
          },
          {
            from: '/other/plans',
            to: '/developers/other/plans',
          },

          // Explore section moved to users/explore
          {
            from: '/explore',
            to: '/users/explore',
          },
          {
            from: '/explore/ai-chat',
            to: '/users/explore/ai-chat',
          },
          {
            from: '/explore/bookmarks',
            to: '/users/explore/bookmarks',
          },
          {
            from: '/explore/exports',
            to: '/users/explore/exports',
          },
          {
            from: '/explore/filters',
            to: '/users/explore/filters',
          },
          {
            from: '/explore/mcp',
            to: '/users/explore/mcp',
          },
          {
            from: '/explore/public-url',
            to: '/users/explore/public-url',
          },
          {
            from: '/explore/time-series',
            to: '/users/explore/time-series',
          },

          // Manage section moved to users/manage
          {
            from: '/manage',
            to: '/users/manage',
          },
          {
            from: '/manage/organization-management',
            to: '/users/manage/organization-management',
          },
          {
            from: '/manage/roles-permissions',
            to: '/users/manage/roles-permissions',
          },
          {
            from: '/manage/user-management',
            to: '/users/manage/user-management',
          },
          {
            from: '/manage/usergroup-management',
            to: '/users/manage/usergroup-management',
          },

          // Specific demo guides that moved to users
          {
            from: '/guides/cost-monitoring-analytics',
            to: '/developers/guides/demos/cost-monitoring-analytics',
          },
          {
            from: '/guides/github-analytics',
            to: '/developers/guides/demos/github-analytics',
          },
          {
            from: '/guides/openrtb-analytics',
            to: '/developers/guides/demos/openrtb-analytics',
          }
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
