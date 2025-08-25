import React from 'react';
import clsx from 'clsx';
import Translate from '@docusaurus/Translate';
import Heading from '@theme/Heading';

export default function NotFoundContent({ className }) {

  return (
    <main className={clsx('container margin-vert--xl', className)}>
      <div className="row">
        <div className="col col--6 col--offset-3">
          <Heading as="h1" className="hero__title">
            <Translate
              id="theme.NotFound.title"
              description="The title of the 404 page">
              Oops!

              The page you were looking for has moved.
            </Translate>
          </Heading>
          <br /><br />

          <img src='/img/404.svg' class='centered' />
          <br />
          <p>
            <Translate
              id="theme.NotFound.p1"
              description="The first paragraph of the 404 page">
              Don't worry! It looks like you've wandered off the beaten path.
              We're here to help you find what you're looking for. Either use the search bar or type Cmd+K to search the docs.
            </Translate>
          </p>

          <p>
            <Translate
              id="theme.NotFound.p2"
              description="The 2nd paragraph of the 404 page">
              You can also explore our documentation using the navigation menu above,
              or search for what you're looking for.
            </Translate>
          </p>

          <p style={{ marginTop: '2rem', fontSize: '0.9rem', opacity: '0.8' }}>
            <Translate
              id="theme.NotFound.reportLink"
              description="Text for reporting broken links">
              Found a broken link?
            </Translate>
            {' '}
            <a
              href="https://github.com/rilldata/rill/issues/new"
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: 'var(--ifm-color-primary)', textDecoration: 'underline' }}
            >
              <Translate
                id="theme.NotFound.reportLinkButton"
                description="Button text for reporting broken links">
                Report it here
              </Translate>
            </a>
          </p>
          <div style={{ textAlign: 'center', margin: '2rem 0' }}>
            <a
              href="/"
              style={{
                display: 'inline-block',
                padding: '0.4rem 0.8rem',
                fontSize: '0.9rem',
                fontWeight: '400',
                textDecoration: 'none',
                borderRadius: '3px',
                transition: 'all 0.2s ease',
                marginRight: '0.5rem',
                border: '1px solid var(--ifm-color-primary)',
                background: 'transparent',
                color: 'var(--ifm-color-primary)',
                cursor: 'pointer'
              }}
            >
              <Translate
                id="theme.NotFound.homeButton"
                description="The home page button text">
                Go to Home Page
              </Translate>
            </a>
          </div>
        </div>
      </div>
    </main >
  );
}
