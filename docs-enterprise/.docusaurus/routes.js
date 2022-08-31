import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/__docusaurus/debug',
    component: ComponentCreator('/__docusaurus/debug', 'abe'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/config',
    component: ComponentCreator('/__docusaurus/debug/config', '79b'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/content',
    component: ComponentCreator('/__docusaurus/debug/content', '910'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/globalData',
    component: ComponentCreator('/__docusaurus/debug/globalData', '3d4'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/metadata',
    component: ComponentCreator('/__docusaurus/debug/metadata', 'fa3'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/registry',
    component: ComponentCreator('/__docusaurus/debug/registry', 'afb'),
    exact: true
  },
  {
    path: '/__docusaurus/debug/routes',
    component: ComponentCreator('/__docusaurus/debug/routes', 'c32'),
    exact: true
  },
  {
    path: '/README',
    component: ComponentCreator('/README', '57e'),
    exact: true
  },
  {
    path: '/',
    component: ComponentCreator('/', '79a'),
    routes: [
      {
        path: '/analyze/embedding-explore',
        component: ComponentCreator('/analyze/embedding-explore', '6ed'),
        exact: true
      },
      {
        path: '/analyze/explore-admin',
        component: ComponentCreator('/analyze/explore-admin', 'f23'),
        exact: true
      },
      {
        path: '/analyze/explore-admin/adding-users',
        component: ComponentCreator('/analyze/explore-admin/adding-users', 'fed'),
        exact: true
      },
      {
        path: '/analyze/explore-admin/adjust-dashboard-layout',
        component: ComponentCreator('/analyze/explore-admin/adjust-dashboard-layout', 'c08'),
        exact: true
      },
      {
        path: '/analyze/explore-admin/admin-security',
        component: ComponentCreator('/analyze/explore-admin/admin-security', 'a9a'),
        exact: true
      },
      {
        path: '/analyze/explore-admin/create-an-external-dashboard',
        component: ComponentCreator('/analyze/explore-admin/create-an-external-dashboard', 'd09'),
        exact: true
      },
      {
        path: '/analyze/explore-admin/explore-json-examples',
        component: ComponentCreator('/analyze/explore-admin/explore-json-examples', 'ebf'),
        exact: true
      },
      {
        path: '/analyze/getting-access-to-rill-dashboards',
        component: ComponentCreator('/analyze/getting-access-to-rill-dashboards', 'eb2'),
        exact: true
      },
      {
        path: '/analyze/getting-started',
        component: ComponentCreator('/analyze/getting-started', 'dc6'),
        exact: true
      },
      {
        path: '/analyze/getting-started/alerting',
        component: ComponentCreator('/analyze/getting-started/alerting', '261'),
        exact: true
      },
      {
        path: '/analyze/getting-started/bookmarking',
        component: ComponentCreator('/analyze/getting-started/bookmarking', '36c'),
        exact: true
      },
      {
        path: '/analyze/getting-started/facet-pivot-table-splits',
        component: ComponentCreator('/analyze/getting-started/facet-pivot-table-splits', '1ba'),
        exact: true
      },
      {
        path: '/analyze/getting-started/scheduled-exports',
        component: ComponentCreator('/analyze/getting-started/scheduled-exports', '5e5'),
        exact: true
      },
      {
        path: '/core-concepts',
        component: ComponentCreator('/core-concepts', '46c'),
        exact: true
      },
      {
        path: '/get-started/data-ingestion-best-practices-1',
        component: ComponentCreator('/get-started/data-ingestion-best-practices-1', '2f5'),
        exact: true
      },
      {
        path: '/get-started/druid-ingestion-optimization',
        component: ComponentCreator('/get-started/druid-ingestion-optimization', '44e'),
        exact: true
      },
      {
        path: '/get-started/metadata-lookups',
        component: ComponentCreator('/get-started/metadata-lookups', '7ca'),
        exact: true
      },
      {
        path: '/get-started/process-streaming-data',
        component: ComponentCreator('/get-started/process-streaming-data', 'dfa'),
        exact: true
      },
      {
        path: '/get-started/process-streaming-data/connecting-with-kafka',
        component: ComponentCreator('/get-started/process-streaming-data/connecting-with-kafka', '622'),
        exact: true
      },
      {
        path: '/get-started/process-streaming-data/real-time-publishing-to-rill',
        component: ComponentCreator('/get-started/process-streaming-data/real-time-publishing-to-rill', '93e'),
        exact: true
      },
      {
        path: '/get-started/process-streaming-data/tutorial-kafka-ingestion',
        component: ComponentCreator('/get-started/process-streaming-data/tutorial-kafka-ingestion', 'e99'),
        exact: true
      },
      {
        path: '/get-started/processing-batch-data',
        component: ComponentCreator('/get-started/processing-batch-data', 'a71'),
        exact: true
      },
      {
        path: '/get-started/processing-batch-data/aws-s3-bucket',
        component: ComponentCreator('/get-started/processing-batch-data/aws-s3-bucket', 'da9'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/get-started/processing-batch-data/azure-storage-container',
        component: ComponentCreator('/get-started/processing-batch-data/azure-storage-container', 'a8d'),
        exact: true
      },
      {
        path: '/get-started/processing-batch-data/gcs-bucket',
        component: ComponentCreator('/get-started/processing-batch-data/gcs-bucket', '56a'),
        exact: true
      },
      {
        path: '/get-started/processing-batch-data/google-bigquery',
        component: ComponentCreator('/get-started/processing-batch-data/google-bigquery', 'fe9'),
        exact: true
      },
      {
        path: '/get-started/processing-batch-data/ingesting-from-big-query',
        component: ComponentCreator('/get-started/processing-batch-data/ingesting-from-big-query', '9b1'),
        exact: true
      },
      {
        path: '/get-started/processing-batch-data/tutorial-druid-ingestion',
        component: ComponentCreator('/get-started/processing-batch-data/tutorial-druid-ingestion', 'ae7'),
        exact: true
      },
      {
        path: '/integrate/authenticating-integrated-applications',
        component: ComponentCreator('/integrate/authenticating-integrated-applications', '94f'),
        exact: true
      },
      {
        path: '/integrate/authenticating-integrated-applications/api-access',
        component: ComponentCreator('/integrate/authenticating-integrated-applications/api-access', '06f'),
        exact: true
      },
      {
        path: '/integrate/authenticating-integrated-applications/api-password',
        component: ComponentCreator('/integrate/authenticating-integrated-applications/api-password', 'fd3'),
        exact: true
      },
      {
        path: '/integrate/authenticating-integrated-applications/jdbc-connection',
        component: ComponentCreator('/integrate/authenticating-integrated-applications/jdbc-connection', '055'),
        exact: true
      },
      {
        path: '/integrate/authenticating-integrated-applications/service-accounts',
        component: ComponentCreator('/integrate/authenticating-integrated-applications/service-accounts', 'dfb'),
        exact: true
      },
      {
        path: '/integrate/jupyter',
        component: ComponentCreator('/integrate/jupyter', 'c93'),
        exact: true
      },
      {
        path: '/integrate/looker',
        component: ComponentCreator('/integrate/looker', '285'),
        exact: true
      },
      {
        path: '/integrate/superset',
        component: ComponentCreator('/integrate/superset', 'd5f'),
        exact: true
      },
      {
        path: '/integrate/tableau',
        component: ComponentCreator('/integrate/tableau', 'e50'),
        exact: true
      },
      {
        path: '/introduction',
        component: ComponentCreator('/introduction', '251'),
        exact: true
      },
      {
        path: '/overview/core-concepts',
        component: ComponentCreator('/overview/core-concepts', '326'),
        exact: true
      },
      {
        path: '/overview/introduction',
        component: ComponentCreator('/overview/introduction', 'bd2'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/resources/contact-support',
        component: ComponentCreator('/resources/contact-support', 'ad7'),
        exact: true
      },
      {
        path: '/resources/faq',
        component: ComponentCreator('/resources/faq', 'b18'),
        exact: true
      },
      {
        path: '/resources/service-status',
        component: ComponentCreator('/resources/service-status', '1ce'),
        exact: true
      },
      {
        path: '/Security/aws-encryption',
        component: ComponentCreator('/Security/aws-encryption', '193'),
        exact: true
      },
      {
        path: '/Security/aws-private-link',
        component: ComponentCreator('/Security/aws-private-link', '1df'),
        exact: true
      },
      {
        path: '/Security/security',
        component: ComponentCreator('/Security/security', 'f86'),
        exact: true
      }
    ]
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
