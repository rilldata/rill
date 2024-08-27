import React from 'react';
import DocCardList from '@theme/DocCardList';
import {useCurrentSidebarCategory} from '@docusaurus/theme-common';

//Using the sidebar to dynamically define the buttons
const CustomDocCardList = (props) => { 
  const category = useCurrentSidebarCategory();

  // List of document ids to exclude
  const excludeIds = ['tutorials/index', 'tutorials/guides',
     'tutorials/guides/index',
     'tutorials/rill_learn_300/overview',
     'tutorials/rill_learn_300/the_end',
      'tutorials/rill_learn_200/rill-cloud',
    'tutorials/rill_learn_200/advanced_developer',
     'tutorials/rill_learn_200/210_0'];

  // Filter out the excluded documents
  const filteredItems = category.items.filter(
    (item) => !excludeIds.includes(item.docId)
  );

  return <DocCardList items={filteredItems} {...props} />;
};

export default CustomDocCardList;
