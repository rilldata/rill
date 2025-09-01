import React from 'react';
import DocSidebar from '@theme-original/DocSidebar';
import { useLocation } from '@docusaurus/router';

export default function DocSidebarWrapper(props) {
  const location = useLocation();
  
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