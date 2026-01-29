import React, { useEffect, useCallback } from 'react';
import DropdownNavbarItem from '@theme-original/NavbarItem/DropdownNavbarItem';
import { useLocation } from '@docusaurus/router';

/**
 * Swizzled DropdownNavbarItem component that closes dropdown on item click.
 * 
 * This fixes the issue where hover-based dropdowns stay open after clicking
 * a menu item because the ::before pseudo-element (hover bridge) keeps the
 * dropdown open if the mouse is still in that area.
 */
export default function DropdownNavbarItemWrapper(props) {
  const location = useLocation();

  // Helper to close all dropdowns and temporarily disable hover
  const closeAllDropdowns = useCallback(() => {
    // Add a class to the navbar to temporarily disable hover
    const navbar = document.querySelector('.navbar');
    if (navbar) {
      navbar.classList.add('navbar--closing-dropdown');
    }

    // Remove dropdown--show class from all dropdowns
    const openDropdowns = document.querySelectorAll('.navbar__item.dropdown.dropdown--show');
    openDropdowns.forEach((dropdown) => {
      dropdown.classList.remove('dropdown--show');
    });

    // Reset aria-expanded attributes and blur
    const expandedLinks = document.querySelectorAll('.navbar__link[aria-expanded="true"]');
    expandedLinks.forEach((link) => {
      link.setAttribute('aria-expanded', 'false');
      link.blur();
    });

    // Remove the closing class after mouse has likely moved away
    setTimeout(() => {
      if (navbar) {
        navbar.classList.remove('navbar--closing-dropdown');
      }
    }, 300);
  }, []);

  // Close dropdown when route changes
  useEffect(() => {
    closeAllDropdowns();
  }, [location.pathname, closeAllDropdowns]);

  // Global click handler for dropdown links
  useEffect(() => {
    const handleClick = (e) => {
      const dropdownLink = e.target.closest('.dropdown__link');
      if (dropdownLink) {
        // Immediately close
        closeAllDropdowns();
      }
    };

    // Use capture phase to catch events before they bubble
    document.addEventListener('click', handleClick, true);
    return () => document.removeEventListener('click', handleClick, true);
  }, [closeAllDropdowns]);

  return <DropdownNavbarItem {...props} />;
}
