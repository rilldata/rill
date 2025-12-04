import React from 'react';
import DocSidebar from '@theme-original/DocSidebar';
import { useLocation } from '@docusaurus/router';

export default function DocSidebarWrapper(props) {
  const location = useLocation();
  const isContactPage = location.pathname.includes('/contact');

  // Add CSS to hide sidebar on contact page
  React.useEffect(() => {
    if (isContactPage) {
      const style = document.createElement('style');
      style.textContent = `
        .theme-doc-sidebar-container {
          display: none !important;
          width: 0 !important;
          min-width: 0 !important;
        }
        .theme-doc-layout {
          grid-template-columns: 1fr !important;
        }
        [class^="docRoot"] {
          --doc-sidebar-width: 0 !important;
        }
      `;
      document.head.appendChild(style);

      return () => {
        document.head.removeChild(style);
      };
    }
  }, [isContactPage]);

  return (
    <div className="custom-doc-sidebar">
      <DocSidebar {...props} />
      <div className="sidebar-release-notes-section">
        <a
          href="/notes"
          className="release-notes-link"
        >
          Release Notes
        </a>
      </div>
    </div>
  );
} 
