# Multi-Panel Dialog Tabs

This directory contains special-purpose tabs used for multi-panel, multi-step dialogs. Currently, these tabs are used in the Create/Edit Alert dialog. 

## Unique Features
These tabs are unique in that they:
- Have a fixed height & width
- Include a number next to each tab
- Are not clickable

## Future Considerations
If these tabs are used in other multi-panel dialogs in the future, consider moving this component set to the `web-common/src/components/dialog` directory.

## Implementation Details

The components in this directory are heavily adapted from the ShadCN's tab component set. The BitsUI tabs, which ShadCN is built upon, have been utilized with the `disabled` prop to create non-clickable tabs. The aesthetics of these disabled tabs have been significantly altered to fit the Alert dialog requirements. 
