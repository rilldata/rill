import { useColorMode } from '@docusaurus/theme-common';
import Navbar from '@theme-original/Navbar';
import { useEffect, useLayoutEffect } from 'react';

// Mobile icon functionality removed

export default function NavbarWrapper(props) {
  // We only need colorMode to determine which icon to show.
  // The toggle logic is handled by the original button's onClick.
  const { colorMode } = useColorMode();

  // Handle Dark Mode Toggle Icons
  // useLayoutEffect fires synchronously before paint, reducing flicker
  useLayoutEffect(() => {
    const toggleButtons = document.querySelectorAll('button[class*="toggleButton"]');
    if (!toggleButtons.length) return;

    toggleButtons.forEach((btn, index) => {
      // Add ID to first toggle for potential targeting
      if (index === 0) btn.id = 'dark-mode-toggle';

      // Find or create the container for our custom icon
      let iconContainer = btn.querySelector('.icon-container');
      if (!iconContainer) {
        iconContainer = document.createElement('span');
        iconContainer.className = 'icon-container';

        // Clear existing Docusaurus toggle content (text/emojis)
        btn.innerHTML = '';
        btn.appendChild(iconContainer);
      }

      // Update Icon based on CURRENT mode (show the OPPOSITE icon)
      // If Dark Mode -> Show Sun (to switch to Light)
      // If Light Mode -> Show Moon (to switch to Dark)
      const isDark = colorMode === 'dark';

      iconContainer.innerHTML = `
        <img 
          src="/icons/${isDark ? 'Sun' : 'Moon'}.svg" 
          alt="${isDark ? 'Switch to light mode' : 'Switch to dark mode'}" 
          width="24" 
          height="24" 
        />
      `;

      btn.setAttribute('aria-label', isDark ? 'Switch to light mode' : 'Switch to dark mode');
    });
  }, [colorMode]);

  // Mobile sidebar icons removed - no longer needed

  // Mark dropdown as active when any dropdown menu item is active
  useEffect(() => {
    const markActiveDropdowns = () => {
      // Target elements with both navbar__item and dropdown classes
      const dropdownItems = document.querySelectorAll('.navbar__item.dropdown');
      dropdownItems.forEach((dropdownItem) => {
        const activeDropdownLink = dropdownItem.querySelector('.dropdown__link--active');
        if (activeDropdownLink) {
          dropdownItem.classList.add('navbar__dropdown--has-active');
        } else {
          dropdownItem.classList.remove('navbar__dropdown--has-active');
        }
      });
    };

    // Check on mount and when DOM changes
    markActiveDropdowns();
    const observer = new MutationObserver(markActiveDropdowns);
    observer.observe(document.body, { childList: true, subtree: true });

    return () => {
      observer.disconnect();
    };
  }, []);

  // Mark non-dropdown navbar items as active when they contain an active link
  useEffect(() => {
    const markActiveNavItems = () => {
      // Target non-dropdown navbar items
      const navItems = document.querySelectorAll('.navbar__item:not(.dropdown)');
      navItems.forEach((navItem) => {
        const activeLink = navItem.querySelector('.navbar__link--active');
        if (activeLink) {
          navItem.classList.add('navbar__item--has-active-link');
        } else {
          navItem.classList.remove('navbar__item--has-active-link');
        }
      });
    };

    // Check on mount and when DOM changes
    markActiveNavItems();
    const observer = new MutationObserver(markActiveNavItems);
    observer.observe(document.body, { childList: true, subtree: true });

    return () => {
      observer.disconnect();
    };
  }, []);

  return <Navbar {...props} />;
}
