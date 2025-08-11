import React from 'react';
import PropTypes from 'prop-types';

/**
 * TileIcon component for documentation tiles.
 * Displays a header, content, and multiple action links.
 * Main card links to demo, with additional links for GitHub and walkthrough.
 */
function TileIcon({ header, content, link, linkLabel = '', target, rel, githubLink, walkthroughLink, icon }) {
    const showArrow = linkLabel !== '';

    return (
        <a className="tile-icon" href={link} target={target} rel={rel}>
            <div className="tile-icon-header">
                {icon && <div className="tile-icon-header-icon-top">{icon}</div>}
                <div className="tile-icon-header-text">{header}</div>
            </div>
            <div className="tile-icon-content">
                {content}
            </div>
            <div className="tile-icon-footer">
                <div className="tile-icon-actions">
                    {githubLink && (
                        <a
                            href={githubLink}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="tile-icon-action-link"
                            onClick={(e) => e.stopPropagation()}
                        >
                            GitHub
                        </a>
                    )}
                    {walkthroughLink && (
                        <a
                            href={walkthroughLink}
                            className="tile-icon-action-link"
                            onClick={(e) => e.stopPropagation()}
                        >
                            Walkthrough
                        </a>
                    )}
                </div>
                <span className={`tile-icon-link-right ${showArrow ? 'with-arrow' : 'no-arrow'}`}>
                    {linkLabel}
                </span>
            </div>
        </a>
    );
}

TileIcon.propTypes = {
    header: PropTypes.string.isRequired,
    content: PropTypes.string.isRequired,
    link: PropTypes.string.isRequired,
    linkLabel: PropTypes.string,
    target: PropTypes.string,
    rel: PropTypes.string,
    githubLink: PropTypes.string,
    walkthroughLink: PropTypes.string,
    icon: PropTypes.node,
};

export default TileIcon; 