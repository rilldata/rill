import React from 'react';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import { translate } from '@docusaurus/Translate';
import { useLocation } from '@docusaurus/router';

export default function HomeBreadcrumbItem() {
    const { pathname } = useLocation();

    // Route the breadcrumb "home" based on the current docs section:
    // - If you're in /developers/*, link to /developers/
    // - If you're in /guide/*, link to /guide/
    // - Otherwise, fall back to site root
    let basePath = '/';
    if (pathname.startsWith('/developers')) {
        basePath = '/';
    } else if (pathname.startsWith('/guide')) {
        basePath = '/guide/';
    }

    const homeHref = useBaseUrl(basePath);

    return (
        <li className="breadcrumbs__item">
            <Link
                aria-label={translate({
                    id: 'theme.docs.breadcrumbs.home',
                    message: 'Home page',
                    description:
                        'The ARIA label for the home page in the breadcrumbs',
                })}
                className="breadcrumbs__link"
                href={homeHref}
            >
                Home
            </Link>
        </li>
    );
}