import React from 'react';

/**
 * Banner directing users to the equivalent page for the other OLAP engine,
 * or indicating this page is only available on one engine.
 *
 * Props:
 *  - engine: "duckdb" | "clickhouse" — which engine THIS page is for
 *  - link: string (optional) — explicit link to the other engine's page
 *  - solo: boolean (default false) — if true, shows "only available on X" (no cross-link)
 *  - message: string (optional) — override the default message text
 */
function WrongOLAP({ engine = 'duckdb', link, solo = false, message }) {
  const isDuckDB = engine === 'duckdb';
  const otherEngine = isDuckDB ? 'ClickHouse' : 'DuckDB';
  const currentEngine = isDuckDB ? 'DuckDB' : 'ClickHouse';
  const logo = isDuckDB
    ? '/img/build/connectors/icons/Logo-DuckDB.svg'
    : '/img/build/connectors/icons/Logo-ClickHouse.svg';

  // Auto-generate link by swapping duckdb↔clickhouse in the current path
  const autoLink = link || (typeof window !== 'undefined'
    ? window.location.pathname.replace(
        isDuckDB ? '/duckdb/' : '/clickhouse/',
        isDuckDB ? '/clickhouse/' : '/duckdb/'
      )
    : '#');

  const crossLinkMessage = (
    <>
      This guide is for <strong>{currentEngine}</strong>.
      Looking for {otherEngine}?{' '}
      <a href={autoLink} style={{ color: 'inherit', textDecoration: 'underline' }}>
        View the {otherEngine} version
      </a>.
    </>
  );

  const soloMessage = (
    <>
      This connector is only available when using <strong>{currentEngine}</strong> as your OLAP engine.
    </>
  );

  const content = message || (solo ? soloMessage : crossLinkMessage);

  return (
    <div className="duckdb-only-banner">
      <div className="duckdb-only-banner__icon">
        <img src={logo} alt={currentEngine} height="20" />
      </div>
      <div className="duckdb-only-banner__content">
        {content}
      </div>
    </div>
  );
}

export default WrongOLAP;
