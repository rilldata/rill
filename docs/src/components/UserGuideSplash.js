import React, { useState, useEffect } from 'react';
import { useHistory } from '@docusaurus/router';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './UserGuideSplash.module.css';

export default function UserGuideSplash() {
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState([]);
    const [isSearching, setIsSearching] = useState(false);
    const history = useHistory();
    const { siteConfig } = useDocusaurusContext();

    // Use Algolia search API from Docusaurus config
    useEffect(() => {
        if (searchQuery.length < 2) {
            setSearchResults([]);
            return;
        }

        setIsSearching(true);

        const timeout = setTimeout(async () => {
            try {
                // Get Algolia config from docusaurus.config.js
                const algoliaConfig = siteConfig.themeConfig?.algolia;

                if (!algoliaConfig) {
                    console.warn('Algolia not configured');
                    setIsSearching(false);
                    return;
                }

                const { appId, apiKey, indexName } = algoliaConfig;

                // Call Algolia search API
                const response = await fetch(
                    `https://${appId}-dsn.algolia.net/1/indexes/${indexName}/query`,
                    {
                        method: 'POST',
                        headers: {
                            'X-Algolia-API-Key': apiKey,
                            'X-Algolia-Application-Id': appId,
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            query: searchQuery,
                            hitsPerPage: 8,
                        }),
                    }
                );

                const data = await response.json();

                if (data.hits && data.hits.length > 0) {
                    const results = data.hits.map(hit => {
                        // Extract just the pathname and hash from the full URL
                        let href = '#';
                        if (hit.url) {
                            try {
                                const url = new URL(hit.url);
                                href = url.pathname + url.hash;
                            } catch (e) {
                                // If URL parsing fails, try to use it as-is
                                href = hit.url;
                            }
                        }

                        return {
                            title: hit.hierarchy?.lvl1 || hit.hierarchy?.lvl0 || hit.content || 'Untitled',
                            subtitle: hit.hierarchy?.lvl2 || hit.hierarchy?.lvl3 || '',
                            href: href,
                            content: hit._snippetResult?.content?.value || '',
                        };
                    });

                    setSearchResults(results);
                } else {
                    setSearchResults([]);
                }
            } catch (error) {
                console.error('Algolia search error:', error);
                setSearchResults([]);
            }

            setIsSearching(false);
        }, 300);

        return () => clearTimeout(timeout);
    }, [searchQuery, siteConfig]);

    const handleSearchChange = (e) => {
        setSearchQuery(e.target.value);
    };

    const handleResultClick = (href) => {
        history.push(href);
    };

    return (
        <div className={styles.splashContainer}>
            <div className={styles.splashContent}>
                <h1 className={styles.splashTitle}>Welcome to User Guides</h1>
                <p className={styles.splashDescription}>
                    Find everything you need to get the most out of Rill.
                    Search our comprehensive guides or browse topics below.
                </p>

                <div className={styles.searchContainer}>
                    <div className={styles.searchInputWrapper}>
                        <svg
                            className={styles.searchIcon}
                            width="20"
                            height="20"
                            viewBox="0 0 20 20"
                            fill="none"
                            xmlns="http://www.w3.org/2000/svg"
                        >
                            <circle
                                cx="8"
                                cy="8"
                                r="6"
                                stroke="currentColor"
                                strokeWidth="2"
                            />
                            <path
                                d="M12.5 12.5l5 5"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeLinecap="round"
                            />
                        </svg>
                        <input
                            type="text"
                            className={styles.searchInput}
                            placeholder="Search user guides..."
                            value={searchQuery}
                            onChange={handleSearchChange}
                            aria-label="Search user guides"
                        />
                        {searchQuery && (
                            <button
                                className={styles.clearButton}
                                onClick={() => setSearchQuery('')}
                                aria-label="Clear search"
                            >
                                ×
                            </button>
                        )}
                    </div>

                    {searchQuery.length >= 2 && (
                        <div className={styles.searchResults}>
                            {isSearching ? (
                                <div className={styles.searchMessage}>Searching...</div>
                            ) : searchResults.length > 0 ? (
                                searchResults.map((result, index) => (
                                    <button
                                        key={index}
                                        className={styles.searchResult}
                                        onClick={() => handleResultClick(result.href)}
                                    >
                                        <span className={styles.resultIcon}>📄</span>
                                        <div className={styles.resultContent}>
                                            <span className={styles.resultTitle}>{result.title}</span>
                                            {result.subtitle && (
                                                <span className={styles.resultSubtitle}>{result.subtitle}</span>
                                            )}
                                        </div>
                                    </button>
                                ))
                            ) : (
                                <div className={styles.searchMessage}>No results found</div>
                            )}
                        </div>
                    )}
                </div>

            </div>
        </div>
    );
}

