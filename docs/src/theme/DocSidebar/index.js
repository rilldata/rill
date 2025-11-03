import React, { useState, useEffect } from 'react';
import DocSidebar from '@theme-original/DocSidebar';
import { useLocation } from '@docusaurus/router';
import { filterSidebarItems } from '../../utils/personaConfig';

export default function DocSidebarWrapper(props) {
  const location = useLocation();
  const isContactPage = location.pathname.includes('/contact');
  const [persona, setPersona] = useState('developer');

  // Listen for persona changes
  useEffect(() => {
    const savedPersona = localStorage.getItem('rill-docs-persona') || 'developer';
    setPersona(savedPersona);

    const handlePersonaChange = (event) => {
      setPersona(event.detail);
    };

    window.addEventListener('persona-change', handlePersonaChange);
    return () => {
      window.removeEventListener('persona-change', handlePersonaChange);
    };
  }, []);

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

  // Filter sidebar items based on persona
  const filteredProps = {
    ...props,
    sidebar: props.sidebar ? filterSidebarItems(props.sidebar, persona) : props.sidebar
  };

  return (
    <div className="custom-doc-sidebar">
      <DocSidebar {...filteredProps} />
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
