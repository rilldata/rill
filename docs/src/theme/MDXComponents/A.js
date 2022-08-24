import React from 'react';
import Link from '@docusaurus/Link';
import Video from '@site/src/components/Video';

export default function MDXA(props) {
  if ((props.href.endsWith(".gif") || props.href.endsWith(".mp4")) && (typeof props.title !== 'undefined')) {
    return (
      <Video vimeoId={props.title} />
    )
  };
  return <Link {...props} />
}
