import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';
import Video from '@site/src/components/Video';

function transformImgClassName(className) {
  return clsx(className, styles.img);
}
export default function MDXImg(props) {
  if (props.src.endsWith(".gif") && (typeof props.title !== 'undefined')) {
    return (
      <Video vimeoId={props.title} />
    )
  };
  return (
    // eslint-disable-next-line jsx-a11y/alt-text
    <img
      loading="lazy"
      {...props}
      className={transformImgClassName(props.className)}
    />
  );
}
