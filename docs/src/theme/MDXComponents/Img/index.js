import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';
import Video from '@site/src/components/Video';

function transformImgClassName(className) {
  return clsx(className, styles.img);
}
export default function MDXImg(props) {
  if ((props.src.endsWith(".gif") || props.src.endsWith(".mp4")) && (typeof props.title !== 'undefined')) {
    return (
      <Video vimeoId={props.title} />
    )
  };
  return (
    <img
      loading="lazy"
      {...props}
      className={transformImgClassName(props.className)}
    />
  );
}
