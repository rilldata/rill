import React from 'react';

/**
 * Beta warning banner for ClickHouse data source pages.
 */
function ClickHousePrereq() {
  return (
    <div
      style={{
        borderLeft: '3px solid var(--ifm-color-emphasis-400)',
        borderRadius: '2px',
        padding: '0.75rem 1rem',
        backgroundColor: 'var(--ifm-color-emphasis-100)',
        fontSize: '0.9rem',
        lineHeight: '1.5',
      }}
    >
      <strong>Beta</strong>
      {' — '}
      Rill-Managed ClickHouse and read-write mode for self-managed ClickHouse are currently in beta.
      For self-managed ClickHouse, we recommend writing to a non-production database, as models
      may drop and recreate tables with matching names.
    </div>
  );
}

export default ClickHousePrereq;
