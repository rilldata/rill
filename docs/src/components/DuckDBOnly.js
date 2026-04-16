import React from 'react';

/**
 * Banner indicating DuckDB-only content.
 *
 * Props:
 *  - toggle: boolean (default true)
 *      true  → "This guide uses DuckDB. Looking for ClickHouse? Jump to the ClickHouse section."
 *      false → "This feature is only available when using DuckDB as your OLAP engine."
 *  - message: string (optional) – override the default message text entirely
 */
function DuckDBOnly({ toggle = true, message }) {
  const handleClick = (e) => {
    e.preventDefault();
    window.location.hash = '#clickhouse';
    const el = document.getElementById('olap-toggle');
    if (el) el.scrollIntoView({ behavior: 'smooth' });
  };

  const toggleMessage = (
    <>
      This guide uses Rill's default embedded engine, <strong>DuckDB</strong>.
      Looking for ClickHouse?{' '}
      <a href="#clickhouse" onClick={handleClick} style={{ color: 'inherit', textDecoration: 'underline' }}>
        Jump to the ClickHouse section
      </a>.
    </>
  );

  const staticMessage = (
    <>
      This feature is only available when using <strong>DuckDB</strong> as your OLAP engine.
    </>
  );

  const content = message || (toggle ? toggleMessage : staticMessage);

  return (
    <div className="duckdb-only-banner">
      <div className="duckdb-only-banner__icon">
        <img
          src="/img/build/connectors/icons/Logo-DuckDB.svg"
          alt="DuckDB"
          height="20"
        />
      </div>
      <div className="duckdb-only-banner__content">
        {content}
      </div>
    </div>
  );
}

export default DuckDBOnly;
