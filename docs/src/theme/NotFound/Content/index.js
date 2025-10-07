import React from 'react';
import clsx from 'clsx';
import Translate from '@docusaurus/Translate';
import Heading from '@theme/Heading';

export default function NotFoundContent({ className }) {

  return (
    <main className={clsx('container margin-vert--xl', className)}>
      <div className="col col--6 col--offset-3" style={{
        textAlign: 'center',
        fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
      }}>
        <img src='/img/404.svg' style={{
          width: '300px',
          marginBottom: '1.5rem'
        }} />
        <Heading as="h3" className="hero__subtitle" style={{
          color: '#1a1a1a',
          fontSize: '1.5rem',
          fontWeight: '600',
          marginBottom: '1rem'
        }}>
          <Translate
            id="theme.NotFound.title"
            description="The title of the 404 page">
            Oops! Page not found
          </Translate>
        </Heading>
        <p style={{
          color: '#71717A',
          fontSize: '1.25rem',
          lineHeight: '1.6',
          marginBottom: '1.5rem'
        }}>
          <Translate
            id="theme.NotFound.p1"
            description="The first paragraph of the 404 page">
            The page you're looking for might have been removed, had its name changed, or is temporarily unavailable.
          </Translate>
        </p>

        <p style={{
          color: '#71717A',
          fontSize: '1.25rem',
          marginBottom: '0'
        }}>
          Return to <a href="/" style={{
            color: '#4736F5',
            textDecoration: 'none',
            fontWeight: '500',
            borderBottom: '2px solid transparent',
            transition: 'all 0.2s ease',
            ':hover': {
              borderBottomColor: '#0070f3'
            }
          }}>Docs</a>
        </p>
      </div>
    </main >
  );
}
