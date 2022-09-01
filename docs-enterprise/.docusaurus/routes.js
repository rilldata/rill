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
    component: ComponentCreator('/', 'd5b'),
    routes: [
      {
        path: '/',
        component: ComponentCreator('/', '70c'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/adding-users',
        component: ComponentCreator('/adding-users', '703'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/admin-security',
        component: ComponentCreator('/admin-security', 'a57'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/alerting',
        component: ComponentCreator('/alerting', '2a1'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/api-access',
        component: ComponentCreator('/api-access', 'f93'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/api-password',
        component: ComponentCreator('/api-password', 'e42'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/authenticating-integrated-applications',
        component: ComponentCreator('/authenticating-integrated-applications', 'ff1'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/aws-encryption',
        component: ComponentCreator('/aws-encryption', '083'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/aws-private-link',
        component: ComponentCreator('/aws-private-link', '2da'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/aws-s3-bucket',
        component: ComponentCreator('/aws-s3-bucket', '353'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/azure-storage-container',
        component: ComponentCreator('/azure-storage-container', 'ce0'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/bookmarking',
        component: ComponentCreator('/bookmarking', 'a01'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/connecting-with-kafka',
        component: ComponentCreator('/connecting-with-kafka', '0cf'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/contact-support',
        component: ComponentCreator('/contact-support', '11f'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/core-concepts',
        component: ComponentCreator('/core-concepts', 'fdb'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/create-an-external-dashboard',
        component: ComponentCreator('/create-an-external-dashboard', 'eba'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/data-ingestion-best-practices-1',
        component: ComponentCreator('/data-ingestion-best-practices-1', '172'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/druid-ingestion-optimization',
        component: ComponentCreator('/druid-ingestion-optimization', '222'),
        exact: true
      },
      {
        path: '/embedding-explore',
        component: ComponentCreator('/embedding-explore', 'e6a'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/explore-admin',
        component: ComponentCreator('/explore-admin', 'f02'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/explore-json-examples',
        component: ComponentCreator('/explore-json-examples', '495'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/facet-pivot-table-splits',
        component: ComponentCreator('/facet-pivot-table-splits', '023'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/faq',
        component: ComponentCreator('/faq', 'c3a'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/gcs-bucket',
        component: ComponentCreator('/gcs-bucket', '324'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/getting-access-to-rill-dashboards',
        component: ComponentCreator('/getting-access-to-rill-dashboards', 'd5a'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/getting-started',
        component: ComponentCreator('/getting-started', '85e'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/google-bigquery',
        component: ComponentCreator('/google-bigquery', 'bfd'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/ingesting-from-big-query',
        component: ComponentCreator('/ingesting-from-big-query', '5b0'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/jdbc-connection',
        component: ComponentCreator('/jdbc-connection', 'e22'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/jupyter',
        component: ComponentCreator('/jupyter', 'bd8'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/looker',
        component: ComponentCreator('/looker', '9fc'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/metadata-lookups',
        component: ComponentCreator('/metadata-lookups', '983'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/process-streaming-data',
        component: ComponentCreator('/process-streaming-data', 'f8a'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/processing-batch-data',
        component: ComponentCreator('/processing-batch-data', '324'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/real-time-publishing-to-rill',
        component: ComponentCreator('/real-time-publishing-to-rill', 'fac'),
        exact: true
      },
      {
        path: '/scheduled-exports',
        component: ComponentCreator('/scheduled-exports', '723'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/security',
        component: ComponentCreator('/security', '2fb'),
        exact: true
      },
      {
        path: '/service-accounts',
        component: ComponentCreator('/service-accounts', '5a8'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/superset',
        component: ComponentCreator('/superset', 'bb4'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/tableau',
        component: ComponentCreator('/tableau', 'dca'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/tutorial-druid-ingestion',
        component: ComponentCreator('/tutorial-druid-ingestion', 'd56'),
        exact: true,
        sidebar: "docsSidebar"
      },
      {
        path: '/tutorial-kafka-ingestion',
        component: ComponentCreator('/tutorial-kafka-ingestion', '0d0'),
        exact: true,
        sidebar: "docsSidebar"
      }
    ]
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
