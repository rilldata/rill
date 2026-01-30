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

  // HubSpot tracking script
  scripts: [
    {
      src: '//js-na2.hs-scripts.com/242088677.js',
      async: true,
      defer: true,
      id: 'hs-script-loader',
    },
  ],

  // Client modules for SPA route tracking
  clientModules: [require.resolve('./src/clientModules/hubspot.js')],

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
            label: "Developers",
            position: "left",
            className: "navbar-docs-link",
            activeBaseRegex: "^(?!/(reference|api|contact|notes|guide)).*", // Keep Docs active for all doc pages

          },
          {
            to: "/guide",
            label: "Guide",
            position: "left",
            className: "navbar-user-guide-link",
            activeBaseRegex: "^/guide.*", // Keep Docs active for all doc pages
          },
          {
            type: "dropdown",
            label: "Reference",
            position: "left",
            to: "/reference/project-files",
            className: 'my-custom-dropdown',
            activeBaseRegex: "^(/reference|/api/admin)",
            items: [
              {
                to: "/reference/project-files",
                label: "Project Files",
              },
              {
                to: "/reference/cli",
                label: "CLI",
              },
              {
                to: "/reference/time-syntax/rill-iso-extensions",
                label: "Rill ISO 8601",
              },
              {
                to: "/api/admin/",
                label: "REST API",
              },

            ],
          },

          // {
          //   type: 'dropdown',
          //   label: 'Reference',
          //   position: 'left',
          //   className: "navbar-reference-link",
          //   items: [
          //     { to: '/reference/project-files', label: 'Project Files' },
          //     { to: '/reference/cli', label: 'CLI' },
          //     { to: '/api/admin/', label: 'REST API' },

          //   ],
          // },
          {
            to: "/contact",
            label: "Contact Us",
            position: "left",
            activeBaseRegex: "^/contact",
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
          // ============================================
          // Legacy paths and misc redirects
          // ============================================
          {
            from: '/install',
            to: '/developers/get-started/install',
          },
          {
            from: '/home/example-repository',
            to: '/',
          },
          {
            from: '/home/install',
            to: '/developers/get-started/install',
          },
          {
            from: '/home/get-started',
            to: '/developers/get-started/quickstart',
          },
          {
            from: '/home/FAQ',
            to: '/developers/other/FAQ',
          },
          {
            from: '/home/concepts/developerVsCloud',
            to: '/developers/deploy/cloud-vs-developer',
          },
          {
            from: '/home/concepts/OLAP',
            to: '/developers/build/connectors/olap#what-is-olap',
          },
          {
            from: '/home/concepts/architecture',
            to: '/developers/get-started/why-rill#architecture',
          },
          {
            from: '/home/concepts/operational',
            to: '/developers/get-started/why-rill#operational-vs-traditional-bi',
          },
          {
            from: '/home/concepts/bi-as-code',
            to: '/developers/get-started/why-rill#bi-as-code',
          },
          {
            from: '/home/concepts/metrics-layer',
            to: '/developers/build/metrics-view',
          },
          {
            from: '/example-projects',
            to: '/#examples',
          },
          {
            from: '/integration/embedding',
            to: '/developers/integrate/embedding',
          },
          {
            from: '/develop/import-data',
            to: '/developers/build/connectors/',
          },
          {
            from: '/develop/sql-models',
            to: '/developers/build/models',
          },
          {
            from: '/develop/metrics-dashboard',
            to: '/developers/build/dashboards',
          },
          {
            from: '/develop/security',
            to: '/developers/build/metrics-view/security',
          },
          {
            from: '/share/user-management',
            to: '/guide/administration/users-and-access/user-management',
          },
          {
            from: '/share/roles-permissions',
            to: '/guide/administration/users-and-access/roles-permissions',
          },
          {
            from: '/share/scheduled-reports',
            to: '/guide/reports/exports',
          },
          {
            from: '/concepts/developerVsCloud',
            to: '/developers/deploy/cloud-vs-developer',
          },
          {
            from: '/concepts/OLAP',
            to: '/developers/build/connectors/olap#what-is-olap',
          },
          {
            from: '/concepts/architecture',
            to: '/developers/get-started/why-rill#architecture',
          },
          {
            from: '/concepts/operational',
            to: '/developers/get-started/why-rill#operational-vs-traditional-bi',
          },
          {
            from: '/concepts/metrics-layer',
            to: '/developers/build/metrics-view',
          },
          {
            from: '/concepts/bi-as-code',
            to: '/developers/get-started/why-rill#bi-as-code',
          },
          {
            from: '/get-started/concepts/cloud-vs-developer',
            to: '/developers/deploy/cloud-vs-developer',
          },
          {
            from: '/get-started/concepts/architecture',
            to: '/developers/get-started/why-rill#architecture',
          },
          {
            from: '/get-started/concepts/operational',
            to: '/developers/get-started/why-rill#operational-vs-traditional-bi',
          },
          {
            from: '/get-started/concepts/bi-as-code',
            to: '/developers/get-started/why-rill#bi-as-code',
          },
          // ============================================
          // /build/* → /developers/build/*
          // ============================================
          {
            from: '/build',
            to: '/developers/build',
          },
          {
            from: '/build/index',
            to: '/developers/build',
          },
          {
            from: '/build/credentials',
            to: '/developers/build/connectors/credentials/',
          },
          {
            from: '/build/ai-configuration',
            to: '/developers/build/ai-configuration',
          },
          {
            from: '/build/project-configuration',
            to: '/developers/build/project-configuration',
          },
          {
            from: '/build/getting-started',
            to: '/developers/build/getting-started',
          },
          {
            from: '/build/structure',
            to: '/developers/build/structure',
          },
          {
            from: '/build/structure/structure',
            to: '/developers/build/structure',
          },
          // Build: Advanced Models
          {
            from: '/build/advanced-models/',
            to: '/developers/build/models/',
          },
          {
            from: '/build/advanced-models/incremental-models',
            to: '/developers/build/models/incremental-models',
          },
          {
            from: '/build/advanced-models/partitions',
            to: '/developers/build/models/partitioned-models',
          },
          {
            from: '/build/advanced-models/staging',
            to: '/developers/build/models/staging-models',
          },
          // Build: Canvas Dashboards
          {
            from: '/build/canvas',
            to: '/developers/build/dashboards/canvas',
          },
          {
            from: '/build/canvas/canvas',
            to: '/developers/build/dashboards/canvas',
          },
          {
            from: '/build/canvas/customization',
            to: '/developers/build/dashboards/customization',
          },
          // Build: Metrics View
          {
            from: '/build/metrics-view',
            to: '/developers/build/metrics-view',
          },
          {
            from: '/build/metrics-view/metrics-view',
            to: '/developers/build/metrics-view',
          },
          {
            from: '/build/metrics-view/customize',
            to: '/developers/build/metrics-view',
          },
          {
            from: '/build/metrics-view/annotations',
            to: '/developers/build/metrics-view/annotations',
          },
          {
            from: '/build/metrics-view/security',
            to: '/developers/build/metrics-view/security',
          },
          {
            from: '/build/metrics-view/time-series',
            to: '/developers/build/metrics-view/time-series',
          },
          {
            from: '/build/metrics-view/underlying-model',
            to: '/developers/build/metrics-view/underlying-model',
          },
          {
            from: '/build/metrics-view/what-are-metrics-views',
            to: '/developers/build/metrics-view/what-are-metrics-views',
          },
          // Build: Metrics View - Dimensions
          {
            from: '/build/metrics-view/dimensions',
            to: '/developers/build/metrics-view/dimensions',
          },
          {
            from: '/build/metrics-view/dimensions/dimensions',
            to: '/developers/build/metrics-view/dimensions',
          },
          {
            from: '/build/metrics-view/dimensions/dimension-uri',
            to: '/developers/build/metrics-view/dimensions/dimension-uri',
          },
          {
            from: '/build/metrics-view/dimensions/lookup',
            to: '/developers/build/metrics-view/dimensions/lookup',
          },
          {
            from: '/build/metrics-view/dimensions/unnesting',
            to: '/developers/build/metrics-view/dimensions/unnesting',
          },
          // Build: Metrics View - Measures
          {
            from: '/build/metrics-view/measures',
            to: '/developers/build/metrics-view/measures',
          },
          {
            from: '/build/metrics-view/measures/measures',
            to: '/developers/build/metrics-view/measures',
          },
          {
            from: '/build/metrics-view/advanced-expressions/advanced-expressions',
            to: '/developers/build/metrics-view/measures',
          },
          {
            from: '/build/metrics-view/advanced-expressions/case-statements',
            to: '/developers/build/metrics-view/measures/case-statements',
          },
          {
            from: '/build/metrics-view/advanced-expressions/fixed-metrics',
            to: '/developers/build/metrics-view/measures/fixed-measures',
          },
          {
            from: '/build/metrics-view/advanced-expressions/metric-formatting',
            to: '/developers/build/metrics-view/measures/measures-formatting',
          },
          {
            from: '/build/metrics-view/advanced-expressions/quantiles',
            to: '/developers/build/metrics-view/measures/quantiles',
          },
          {
            from: '/build/metrics-view/advanced-expressions/referencing',
            to: '/developers/build/metrics-view/measures/referencing',
          },
          {
            from: '/build/metrics-view/advanced-expressions/windows',
            to: '/developers/build/metrics-view/measures/windows',
          },
          // Build: Models
          {
            from: '/build/models',
            to: '/developers/build/models',
          },
          {
            from: '/build/models/index',
            to: '/developers/build/models',
          },
          {
            from: '/build/models/data-quality-tests',
            to: '/developers/build/models/data-quality-tests',
          },
          {
            from: '/build/models/data-refresh',
            to: '/developers/build/models/data-refresh',
          },
          {
            from: '/build/models/incremental-models',
            to: '/developers/build/models/incremental-models',
          },
          {
            from: '/build/models/incremental-partitioned-models',
            to: '/developers/build/models/incremental-partitioned-models',
          },
          {
            from: '/build/models/model-differences',
            to: '/developers/build/models/model-differences',
          },
          {
            from: '/build/models/models-101',
            to: '/developers/build/models/models-101',
          },
          {
            from: '/build/models/partitioned-models',
            to: '/developers/build/models/partitioned-models',
          },
          {
            from: '/build/models/performance',
            to: '/developers/build/models/performance',
          },
          {
            from: '/build/models/source-models',
            to: '/developers/build/models/source-models',
          },
          {
            from: '/build/models/sql-models',
            to: '/developers/build/models/sql-models',
          },
          {
            from: '/build/models/staging-models',
            to: '/developers/build/models/staging-models',
          },
          {
            from: '/build/models/templating',
            to: '/developers/build/models/templating',
          },
          // Build: Dashboards
          {
            from: '/build/dashboards',
            to: '/developers/build/dashboards',
          },
          {
            from: '/build/dashboards/dashboards',
            to: '/developers/build/dashboards',
          },
          {
            from: '/build/dashboards/canvas',
            to: '/developers/build/dashboards/canvas',
          },
          {
            from: '/build/dashboards/customization',
            to: '/developers/build/dashboards/customization',
          },
          {
            from: '/build/dashboards/dashboards-101',
            to: '/developers/build/dashboards/dashboards-101',
          },
          {
            from: '/build/dashboards/explore',
            to: '/developers/build/dashboards/explore',
          },
          {
            from: '/build/dashboards/canvas-widgets',
            to: '/developers/build/dashboards/canvas-widgets',
          },
          {
            from: '/build/dashboards/canvas-widgets/canvas-widgets',
            to: '/developers/build/dashboards/canvas-widgets',
          },
          {
            from: '/build/dashboards/canvas-widgets/chart',
            to: '/developers/build/dashboards/canvas-widgets/chart',
          },
          {
            from: '/build/dashboards/canvas-widgets/data',
            to: '/developers/build/dashboards/canvas-widgets/data',
          },
          {
            from: '/build/dashboards/canvas-widgets/misc',
            to: '/developers/build/dashboards/canvas-widgets/misc',
          },
          // Build: Connectors
          {
            from: '/build/connectors',
            to: '/developers/build/connectors',
          },
          {
            from: '/build/connectors/connectors',
            to: '/developers/build/connectors',
          },
          {
            from: '/build/connectors/credentials',
            to: '/developers/build/connectors/credentials',
          },
          {
            from: '/build/connectors/templating',
            to: '/developers/build/connectors/templating',
          },
          {
            from: '/build/connectors/data-source',
            to: '/developers/build/connectors/data-source',
          },
          {
            from: '/build/connectors/data-source/data-source',
            to: '/developers/build/connectors/data-source',
          },
          {
            from: '/build/connectors/data-source/athena',
            to: '/developers/build/connectors/data-source/athena',
          },
          {
            from: '/build/connectors/data-source/azure',
            to: '/developers/build/connectors/data-source/azure',
          },
          {
            from: '/build/connectors/data-source/bigquery',
            to: '/developers/build/connectors/data-source/bigquery',
          },
          {
            from: '/build/connectors/data-source/duckdb',
            to: '/developers/build/connectors/data-source/duckdb',
          },
          {
            from: '/build/connectors/data-source/gcs',
            to: '/developers/build/connectors/data-source/gcs',
          },
          {
            from: '/build/connectors/data-source/googlesheets',
            to: '/developers/build/connectors/data-source/googlesheets',
          },
          {
            from: '/build/connectors/data-source/https',
            to: '/developers/build/connectors/data-source/https',
          },
          {
            from: '/build/connectors/data-source/kafka',
            to: '/developers/build/connectors/data-source/kafka',
          },
          {
            from: '/build/connectors/data-source/local-file',
            to: '/developers/build/connectors/data-source/local-file',
          },
          {
            from: '/build/connectors/data-source/mysql',
            to: '/developers/build/connectors/data-source/mysql',
          },
          {
            from: '/build/connectors/data-source/openai',
            to: '/developers/build/connectors/ai/openai',
          },
          {
            from: '/build/connectors/data-source/postgres',
            to: '/developers/build/connectors/data-source/postgres',
          },
          {
            from: '/build/connectors/data-source/redshift',
            to: '/developers/build/connectors/data-source/redshift',
          },
          {
            from: '/build/connectors/data-source/s3',
            to: '/developers/build/connectors/data-source/s3',
          },
          {
            from: '/build/connectors/data-source/salesforce',
            to: '/developers/build/connectors/data-source/salesforce',
          },
          {
            from: '/build/connectors/data-source/slack',
            to: '/developers/build/connectors/data-source/slack',
          },
          {
            from: '/build/connectors/data-source/snowflake',
            to: '/developers/build/connectors/data-source/snowflake',
          },
          {
            from: '/build/connectors/data-source/sqlite',
            to: '/developers/build/connectors/data-source/sqlite',
          },
          {
            from: '/build/connectors/olap',
            to: '/developers/build/connectors/olap',
          },
          {
            from: '/build/connectors/olap/olap',
            to: '/developers/build/connectors/olap',
          },
          {
            from: '/build/connectors/olap/clickhouse',
            to: '/developers/build/connectors/olap/clickhouse',
          },
          {
            from: '/build/connectors/olap/druid',
            to: '/developers/build/connectors/olap/druid',
          },
          {
            from: '/build/connectors/olap/duckdb',
            to: '/developers/build/connectors/olap/duckdb',
          },
          {
            from: '/build/connectors/olap/motherduck',
            to: '/developers/build/connectors/olap/motherduck',
          },
          {
            from: '/build/connectors/olap/multiple-olap',
            to: '/developers/build/connectors/olap/multiple-olap',
          },
          {
            from: '/build/connectors/olap/pinot',
            to: '/developers/build/connectors/olap/pinot',
          },
          // Build: Custom APIs
          {
            from: '/build/custom-apis',
            to: '/developers/build/custom-apis',
          },
          {
            from: '/build/custom-apis/custom-apis',
            to: '/developers/build/custom-apis',
          },
          // Build: Debugging
          {
            from: '/build/debugging',
            to: '/developers/build/debugging',
          },
          {
            from: '/build/debugging/index',
            to: '/developers/build/debugging',
          },
          {
            from: '/build/debugging/trace-viewer',
            to: '/developers/build/debugging/trace-viewer',
          },
          // Build: IDE
          {
            from: '/build/ide',
            to: '/developers/build/ide',
          },
          {
            from: '/build/ide/ide',
            to: '/developers/build/ide',
          },
          // ============================================
          // /build/connect/* → /developers/build/connectors/*
          // ============================================
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
            to: '/developers/build/connectors/ai/openai',
          },
          // ============================================
          // /connect/* → /developers/build/connectors/*
          // ============================================
          {
            from: '/connect',
            to: '/developers/build/connectors',
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
            to: '/developers/build/connectors/ai/openai',
          },
          // ============================================
          // /deploy/* → /developers/deploy/*
          // ============================================
          {
            from: '/deploy',
            to: '/developers/deploy',
          },
          {
            from: '/deploy/index',
            to: '/developers/deploy',
          },
          {
            from: '/deploy/cloud-vs-developer',
            to: '/developers/deploy/cloud-vs-developer',
          },
          {
            from: '/deploy/deploy-credentials',
            to: '/developers/deploy/deploy-credentials',
          },
          {
            from: '/deploy/deploy-dashboard',
            to: '/developers/deploy/deploy-dashboard',
          },
          {
            from: '/deploy/deploy-dashboard/deploy-dashboard',
            to: '/developers/deploy/deploy-dashboard',
          },
          {
            from: '/deploy/deploy-dashboard/deploy-from-cli',
            to: '/developers/deploy/deploy-dashboard/deploy-from-cli',
          },
          {
            from: '/deploy/deploy-dashboard/github-101',
            to: '/developers/deploy/deploy-dashboard/github-101',
          },
          {
            from: '/deploy/project-errors',
            to: '/developers/deploy/project-errors',
          },
          {
            from: '/deploy/performance',
            to: '/developers/guides/performance',
          },
          {
            from: '/deploy/templating',
            to: '/developers/build/connectors/templating',
          },
          {
            from: '/deploy/source-refresh',
            to: '/developers/build/models/data-refresh',
          },
          {
            from: '/deploy/credentials/',
            to: '/developers/build/connectors/credentials/',
          },
          {
            from: '/deploy/credentials/s3',
            to: '/developers/build/connectors/data-source/s3',
          },
          {
            from: '/deploy/credentials/gcs',
            to: '/developers/build/connectors/data-source/gcs',
          },
          {
            from: '/deploy/credentials/azure',
            to: '/developers/build/connectors/data-source/azure',
          },
          {
            from: '/deploy/credentials/athena',
            to: '/developers/build/connectors/data-source/athena',
          },
          {
            from: '/deploy/credentials/bigquery',
            to: '/developers/build/connectors/data-source/bigquery',
          },
          {
            from: '/deploy/credentials/snowflake',
            to: '/developers/build/connectors/data-source/snowflake',
          },
          {
            from: '/deploy/credentials/postgres',
            to: '/developers/build/connectors/data-source/postgres',
          },
          {
            from: '/deploy/credentials/salesforce',
            to: '/developers/build/connectors/data-source/salesforce',
          },
          {
            from: '/deploy/credentials/motherduck',
            to: '/developers/build/connectors/olap/motherduck',
          },
          // ============================================
          // /get-started/* → /developers/get-started/*
          // ============================================
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
          {
            from: '/get-started/quickstart',
            to: '/developers/get-started/quickstart',
          },
          {
            from: '/get-started/quickstart/quickstart',
            to: '/developers/get-started/quickstart',
          },
          {
            from: '/get-started/why-rill',
            to: '/developers/get-started/why-rill',
          },
          // ============================================
          // /guides/* → /developers/guides/*
          // ============================================
          {
            from: '/guides',
            to: '/developers/guides',
          },
          {
            from: '/guides/index',
            to: '/developers/guides',
          },
          {
            from: '/guides/clone-a-project',
            to: '/developers/guides/clone-a-project',
          },
          {
            from: '/guides/cost-monitoring-analytics',
            to: '/developers/guides/cost-monitoring-analytics',
          },
          {
            from: '/guides/github-analytics',
            to: '/developers/guides/github-analytics',
          },
          {
            from: '/guides/integrating-with-rill',
            to: '/developers/guides/integrating-with-rill',
          },
          {
            from: '/guides/openrtb-analytics',
            to: '/developers/guides/openrtb-analytics',
          },
          {
            from: '/guides/performance',
            to: '/developers/guides/performance',
          },
          {
            from: '/guides/setting-up-mcp',
            to: '/developers/guides/setting-up-mcp',
          },
          {
            from: '/guides/rill-basics',
            to: '/developers/guides/rill-basics/launch',
          },
          {
            from: '/guides/rill-basics/1-launch',
            to: '/developers/guides/rill-basics/launch',
          },
          {
            from: '/guides/rill-basics/2-import',
            to: '/developers/guides/rill-basics/import',
          },
          {
            from: '/guides/rill-basics/3-model',
            to: '/developers/guides/rill-basics/model',
          },
          {
            from: '/guides/rill-basics/4-metrics-view',
            to: '/developers/guides/rill-basics/metrics-view',
          },
          {
            from: '/guides/rill-basics/5-dashboard',
            to: '/developers/guides/rill-basics/dashboard',
          },
          {
            from: '/guides/rill-basics/6-deploy',
            to: '/developers/guides/rill-basics/deploy',
          },
          {
            from: '/guides/rill-basics/success',
            to: '/developers/guides/rill-basics/success',
          },
          {
            from: '/guides/rill-clickhouse',
            to: '/developers/guides/rill-clickhouse',
          },
          {
            from: '/guides/rill-clickhouse/index',
            to: '/developers/guides/rill-clickhouse',
          },
          {
            from: '/guides/rill-clickhouse/1-r_ch_launch',
            to: '/developers/guides/rill-clickhouse/r_ch_launch',
          },
          {
            from: '/guides/rill-clickhouse/2-r_ch_connect',
            to: '/developers/guides/rill-clickhouse/r_ch_connect',
          },
          {
            from: '/guides/rill-clickhouse/3-r_ch_metrics-view',
            to: '/developers/guides/rill-clickhouse/r_ch_metrics-view',
          },
          {
            from: '/guides/rill-clickhouse/4-r_ch_dashboard',
            to: '/developers/guides/rill-clickhouse/r_ch_dashboard',
          },
          {
            from: '/guides/rill-clickhouse/5-r_ch_deploy',
            to: '/developers/guides/rill-clickhouse/r_ch_deploy',
          },
          {
            from: '/guides/rill-clickhouse/r_ch_ingest',
            to: '/developers/guides/rill-clickhouse/r_ch_ingest',
          },
          // ============================================
          // /integrate/* → /developers/integrate/*
          // ============================================
          {
            from: '/integrate',
            to: '/developers/integrate',
          },
          {
            from: '/integrate/index',
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
          {
            from: '/integrate/url-parameters',
            to: '/developers/integrate/url-parameters',
          },
          {
            from: '/integrate/custom-apis',
            to: '/developers/build/custom-apis',
          },
          {
            from: '/integrate/custom-apis/metrics-sql-api',
            to: '/developers/build/custom-apis',
          },
          {
            from: '/integrate/custom-apis/sql-api',
            to: '/developers/build/custom-apis',
          },
          // ============================================
          // /other/* → /developers/other/*
          // ============================================
          {
            from: '/other/FAQ',
            to: '/developers/other/FAQ',
          },
          {
            from: '/other/plans',
            to: '/developers/other/plans',
          },
          {
            from: '/other/v50-dashboard-changes',
            to: '/developers/other/v50-dashboard-changes',
          },
          {
            from: '/other/granting',
            to: '/developers/other/granting',
          },
          {
            from: '/other/granting/granting',
            to: '/developers/other/granting',
          },
          {
            from: '/other/granting/aws-s3-bucket',
            to: '/developers/other/granting/aws-s3-bucket',
          },
          {
            from: '/other/granting/azure-storage-container',
            to: '/developers/other/granting/azure-storage-container',
          },
          {
            from: '/other/granting/gcs-bucket',
            to: '/developers/other/granting/gcs-bucket',
          },
          {
            from: '/other/granting/google-bigquery',
            to: '/developers/other/granting/google-bigquery',
          },
          // ============================================
          // /manage/* → /guide/administration/*
          // ============================================
          {
            from: '/manage',
            to: '/guide/administration',
          },
          {
            from: '/manage/index',
            to: '/guide/administration',
          },
          {
            from: '/manage/security',
            to: '/developers/build/metrics-view/security',
          },
          {
            from: '/manage/organization-management',
            to: '/guide/administration/organization-settings',
          },
          {
            from: '/manage/project-management',
            to: '/guide/administration/project-settings',
          },
          {
            from: '/manage/project-management/variables-and-credentials',
            to: '/guide/administration/project-settings/variables-and-credentials',
          },
          {
            from: '/manage/roles-permissions',
            to: '/guide/administration/users-and-access/roles-permissions',
          },
          {
            from: '/manage/user-management',
            to: '/guide/administration/users-and-access/user-management',
          },
          {
            from: '/manage/usergroup-management',
            to: '/guide/administration/users-and-access/usergroup-management',
          },
          {
            from: '/manage/service-tokens',
            to: '/guide/administration/access-tokens/service-tokens',
          },
          {
            from: '/manage/user-tokens',
            to: '/guide/administration/access-tokens/user-tokens',
          },
          {
            from: '/manage/account-management/billing',
            to: '/developers/other/plans',
          },
          {
            from: '/manage/granting/',
            to: '/developers/other/granting/',
          },
          {
            from: '/manage/granting/aws-s3-bucket',
            to: '/developers/other/granting/aws-s3-bucket',
          },
          {
            from: '/manage/granting/azure-storage-container',
            to: '/developers/other/granting/azure-storage-container',
          },
          {
            from: '/manage/granting/gcs-bucket',
            to: '/developers/other/granting/gcs-bucket',
          },
          {
            from: '/manage/granting/google-bigquery',
            to: '/developers/other/granting/google-bigquery',
          },
          // ============================================
          // /explore/* → /guide/*
          // ============================================
          {
            from: '/explore',
            to: '/guide/dashboards',
          },
          {
            from: '/explore/index',
            to: '/guide/dashboards',
          },
          {
            from: '/explore/ai-chat',
            to: '/guide/ai/ai-chat',
          },
          {
            from: '/explore/mcp',
            to: '/guide/ai/mcp',
          },
          {
            from: '/explore/alerts',
            to: '/guide/alerts',
          },
          {
            from: '/explore/bookmarks',
            to: '/guide/dashboards/bookmarks',
          },
          {
            from: '/explore/canvas',
            to: '/guide/dashboards/canvas',
          },
          {
            from: '/explore/exports',
            to: '/guide/reports/exports',
          },
          {
            from: '/explore/filters',
            to: '/guide/dashboards/filters',
          },
          {
            from: '/explore/filters/filters',
            to: '/guide/dashboards/filters',
          },
          {
            from: '/explore/filters/time-series',
            to: '/guide/dashboards/time-series',
          },
          {
            from: '/explore/public-url',
            to: '/guide/dashboards/public-urls',
          },
          {
            from: '/explore/time-series',
            to: '/guide/dashboards/time-series',
          },
          {
            from: '/explore/dashboard-101',
            to: '/guide/dashboards/explore',
          },
          {
            from: '/explore/dashboard-101/dashboard-101',
            to: '/guide/dashboards/explore',
          },
          {
            from: '/explore/dashboard-101/multi-metrics',
            to: '/guide/dashboards/explore/multi-metrics',
          },
          {
            from: '/explore/dashboard-101/pivot',
            to: '/guide/dashboards/explore/pivot',
          },
          {
            from: '/explore/dashboard-101/tdd',
            to: '/guide/dashboards/explore/tdd',
          },
          // ============================================
          // /reference/* → /developers/build/* or /reference/*
          // ============================================
          {
            from: '/reference/templating',
            to: '/developers/build/connectors/templating',
          },
          {
            from: '/reference/project-files/advanced-models',
            to: '/reference/project-files/models',
          },
          {
            from: '/reference/rill-iso-extensions',
            to: '/reference/time-syntax/rill-iso-extensions',
          },
          {
            from: '/reference/olap-engines/',
            to: '/developers/build/connectors/olap/',
          },
          {
            from: '/reference/olap-engines/duckdb',
            to: '/developers/build/connectors/olap/duckdb',
          },
          {
            from: '/reference/olap-engines/clickhouse',
            to: '/developers/build/connectors/olap/clickhouse',
          },
          {
            from: '/reference/olap-engines/druid',
            to: '/developers/build/connectors/olap/druid',
          },
          {
            from: '/reference/olap-engines/pinot',
            to: '/developers/build/connectors/olap/pinot',
          },
          {
            from: '/reference/olap-engines/multiple-olap',
            to: '/developers/build/connectors/olap/multiple-olap',
          },
          {
            from: '/reference/connectors/',
            to: '/developers/build/connectors/',
          },
          {
            from: '/reference/connectors/gcs',
            to: '/developers/build/connectors/data-source/gcs',
          },
          {
            from: '/reference/connectors/azure',
            to: '/developers/build/connectors/data-source/azure',
          },
          {
            from: '/reference/connectors/s3',
            to: '/developers/build/connectors/data-source/s3',
          },
          {
            from: '/reference/connectors/snowflake',
            to: '/developers/build/connectors/data-source/snowflake',
          },
          {
            from: '/reference/connectors/bigquery',
            to: '/developers/build/connectors/data-source/bigquery',
          },
          {
            from: '/reference/connectors/redshift',
            to: '/developers/build/connectors/data-source/redshift',
          },
          {
            from: '/reference/connectors/postgres',
            to: '/developers/build/connectors/data-source/postgres',
          },
          {
            from: '/reference/connectors/athena',
            to: '/developers/build/connectors/data-source/athena',
          },
          {
            from: '/reference/connectors/mysql',
            to: '/developers/build/connectors/data-source/mysql',
          },
          {
            from: '/reference/connectors/sqlite',
            to: '/developers/build/connectors/data-source/sqlite',
          },
          {
            from: '/reference/connectors/salesforce',
            to: '/developers/build/connectors/data-source/salesforce',
          },
          {
            from: '/reference/connectors/sheets',
            to: '/developers/build/connectors/data-source/googlesheets',
          },
          {
            from: '/reference/connectors/slack',
            to: '/developers/build/connectors/data-source/slack',
          },
          {
            from: '/reference/connectors/local-file',
            to: '/developers/build/connectors/data-source/local-file',
          },
          {
            from: '/reference/connectors/https',
            to: '/developers/build/connectors/data-source/https',
          },
          // ============================================
          // /developers/build/connectors/data-source/<ai> -> /developers/build/connectors/ai/<ai>
          // ============================================
          {
            from: '/developers/build/connectors/data-source/openai',
            to: '/developers/build/connectors/ai/openai',
          }


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
