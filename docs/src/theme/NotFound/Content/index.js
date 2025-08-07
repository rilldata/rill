import React, { useEffect } from 'react';
import clsx from 'clsx';
import Translate from '@docusaurus/Translate';
import Heading from '@theme/Heading';

export default function NotFoundContent({ className }) {
  useEffect(() => {
    // Auto-redirect to home page after 3 seconds
    const timer = setTimeout(() => {
      window.location.href = '/';
    }, 5000);

    return () => clearTimeout(timer);
  }, []);

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
              We'll redirect you back to the home page in just a moment.
            </Translate>
          </p>
          <p>
            <Translate
              id="theme.NotFound.p2"
              description="The 2nd paragraph of the 404 page">
              If you'd like to explore our documentation, you can start from the
            </Translate>
            {' '}
            <a href="/" style={{ color: 'var(--ifm-color-primary)', textDecoration: 'underline' }}>
              home page
            </a>
            {' '}
            <Translate
              id="theme.NotFound.p2_continued"
              description="The continuation of the 2nd paragraph of the 404 page">
              or use the navigation menu above.
            </Translate>
          </p>
          <p style={{ fontSize: '0.9em', color: 'var(--ifm-color-emphasis-600)', fontStyle: 'italic' }}>
            Redirecting automatically in 5 seconds...
          </p>
        </div>
      </div>
    </main>
  );
}
