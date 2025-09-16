---
title: Canvas Components
description: Complete guide to all available components in Rill Canvas Dashboards
---

import ComponentTile from '@site/src/components/ComponentTile';

Rill Canvas dashboards are built using a variety of components that can display data, create visualizations, and add rich content. Each component can be created dynamically through the visual Canvas dashboard editor or defined in individual YAML files. For more information, refer to our [Components reference doc](/reference/project-files/component).


## All Components

### Data 

<div className="component-icon-grid">
    <ComponentTile
        header="KPIs"
        link="/build/dashboards/canvas-components/data#kpi-grid"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/kpi.png" alt="KPI" />}
    />
    <ComponentTile
        header="Leaderboard"
        link="/build/dashboards/canvas-components/data#leaderboard"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/leaderboard.png" alt="Leaderboard" />}
    />
    <ComponentTile
        header="Pivot / Table"
        link="/build/dashboards/canvas-components/data#pivottable"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/table.png" alt="Table" />}
    />
</div>

### Chart 

<div className="component-icon-grid">
    <ComponentTile
        header="Bar" 
        link="/build/dashboards/canvas-components/chart#bar-chart"
        multiple_measures="True"
        image={<img src="/img/build/canvas/components/bar.png" alt="Bar Chart" />}
    />
    <ComponentTile
        header="Line"
        link="/build/dashboards/canvas-components/chart#line-chart"
        multiple_measures="True"
        image={<img src="/img/build/canvas/components/line.png" alt="Line Chart" />}
    />
    <ComponentTile
        header="Stacked Area"
        link="/build/dashboards/canvas-components/chart#stacked-area-chart"
        multiple_measures="True"
        image={<img src="/img/build/canvas/components/stacked-area.png" alt="Stacked Area Chart" />}
    />
    <ComponentTile
        header="Stacked Bar"
        link="/build/dashboards/canvas-components/chart#stacked-bar-chart"
        multiple_measures="True"
        image={<img src="/img/build/canvas/components/stacked-bar.png" alt="Stacked Bar Chart" />}
    />
    <ComponentTile
        header="Stacked Bar Normalized"
        link="/build/dashboards/canvas-components/chart#stacked-bar-normalized"
        multiple_measures="True"
        image={<img src="/img/build/canvas/components/stacked-bar-normalized.png" alt="Stacked Bar Normalized Chart" />}
    />
    <ComponentTile
        header="Donut"
        link="/build/dashboards/canvas-components/chart#donut-chart"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/donut.png" alt="Donut Chart" />}
    />
    <ComponentTile
        header="Funnel"
        link="/build/dashboards/canvas-components/chart#funnel-chart"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/funnel.png" alt="Funnel Chart" />}
    />
    <ComponentTile
        header="Heat Map"
        link="/build/dashboards/canvas-components/chart#heat-map"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/heatmap.png" alt="Heat Map" />}
    />
    <ComponentTile
        header="Combo"
        link="/build/dashboards/canvas-components/chart#combo-chart"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/combo.png" alt="Combo Chart" />}
    />
</div>

### Miscellaneous 

<div className="component-icon-grid">
    <ComponentTile
        header="Text"
        link="/build/dashboards/canvas-components/misc#textmarkdown"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/text.png" alt="Text" />}
    />
    <ComponentTile
        header="Image"
        link="/build/dashboards/canvas-components/misc#image"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/image.png" alt="Image" />}
    /> 
    <ComponentTile
        header="Component"
        link="/build/dashboards/canvas-components/data#component"
        multiple_measures="False"
        image={<img src="/img/build/canvas/components/component.png" alt="Component" />}
    />
</div>

