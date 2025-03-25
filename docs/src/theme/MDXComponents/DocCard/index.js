import React from 'react';
import Link from '@docusaurus/Link';
import styles from './styles.module.scss'; // rename if needed

const DocCard = ({ item }) => {
    return (
        <Link to={item.href} className={styles.card}>
            <div className={styles.header}>
                <h3>{item.label}</h3>
            </div>
        </Link>
    );
};

export default DocCard;
