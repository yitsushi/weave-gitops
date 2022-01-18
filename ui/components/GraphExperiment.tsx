import * as d3 from "d3";
import * as React from "react";
import styled from "styled-components";

type Props = {
  className?: string;
};

function GraphExperiment({ className }: Props) {
  const data = {
    nodes: [{ id: 0 }, { id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }, { id: 5 }],
    links: [
      { source: 0, target: 1 },
      { source: 0, target: 2 },
      { source: 0, target: 3 },
      { source: 3, target: 4 },
      { source: 3, target: 5 },
    ],
  };

  const Graph = function constructor(opts) {
    const element = opts.element || {},
      margin = opts.margin || 20,
      width = opts.width || element.offsetWidth || 900,
      height = (opts.height || element.offsetHeight || 600) - 0.5 - margin,
      distance = opts.distance || 0.5e2,
      radius = opts.radius || 5e1,
      strength = opts.strength || -8e3,
      graph = opts.graph || {},
      onSelection =
        opts.onSelection ||
        function () {
          ("");
        };
    this.scale = (opts.dragAndZoom && opts.dragAndZoom.scale) || [0.1, 4];

    // SVG
    const svg = d3
      .select(element)
      .append("svg")
      .attr("width", width)
      .attr("height", height);

    const g = svg.append("g").attr("class", "graph-container");

    // Links
    const link = g
      .append("g")
      .attr("class", "links")
      .selectAll("line")
      .data(graph.links)
      .enter()
      .append("line");

    // Nodes
    const node = g
      .append("g")
      .attr("class", "nodes")
      .selectAll("g")
      .data(graph.nodes)
      .enter()
      .append("g");

    node
      .append("circle")
      .attr("r", radius)
      .classed("highlight", function (d) {
        return d.type === "hub";
      });

    // node.append('text')
    //   .text(function(d) { return d.id; });

    // ------------------
    // Force Simulation
    // ------------------
    const simulation = d3
      .forceSimulation()
      .force(
        "link",
        d3
          .forceLink()
          .id(function (d) {
            return d.id;
          })
          .iterations(4)
      )
      // .distance(distance))
      .force("charge", d3.forceManyBody().strength(strength))
      .force("center", d3.forceCenter(width / 2, height / 2))
      .force("x", d3.forceX())
      .force("y", d3.forceY());

    simulation.nodes(graph.nodes).on("tick", function ticked() {
      link
        .attr("x1", function (d) {
          return d.source.x;
        })
        .attr("y1", function (d) {
          return d.source.y;
        })
        .attr("x2", function (d) {
          return d.target.x;
        })
        .attr("y2", function (d) {
          return d.target.y;
        });

      node.attr("transform", function (d) {
        return "translate(" + d.x + "," + d.y + ")";
      });
    });

    simulation.force("link").links(graph.links);

    // ----- /Simulation

    this.svg = svg;
    this.g = g;
    this.nodes = node;
    this.links = link;
    this.data = graph;
    this.simulation = simulation;
    this.onSelection = onSelection;
    this.height = height;
    this.width = width;

    if (opts.highlightSelectedPath) {
      this.highlightSelectedPath();
    }

    if (opts.dragAndZoom) {
      this.dragAndZoom();
    }
  };

  // ------------------
  // Zoom
  // ------------------
  Graph.prototype.dragAndZoom = function () {
    const zoom = d3
      .zoom()
      .scaleExtent(this.scale)
      // .scaleExtent([0.5, 10])
      // .translateExtent([[0, -800], [600, 800]])
      .on("zoom", function zoomed(e) {
        console.log(this.g);
        // console.log(e.transform);
        // console.log(e.scale, e.translate) // v3
        this.g.attr("transform", e.transform);
      });

    this.svg.call(zoom);

    // reset button
    d3.select(".reset").on("click", function resetted() {
      this.svg.transition().duration(750).call(zoom.transform, d3.zoomIdentity);
    });
  };
  const divRef = React.useRef();

  React.useEffect(() => {
    if (!divRef.current) return;
    else {
      const graph = new Graph({
        element: divRef.current,
        width: 1100,
        height: 600,
        strength: -50, // -8e1
        distance: 6,
        radius: 5,
        graph: data,
        dragAndZoom: { scale: [0.1, 2] },
      });
    }
  }, [divRef]);
  return <div className={className} ref={divRef} id="graph"></div>;
}

export default styled(GraphExperiment)`
  .links line {
    stroke-width: 1;
    stroke: white;
    stroke: #049fd9;
  }

  .nodes circle {
    stroke-width: 2;
    stroke: white;
    stroke: #049fd9;
    fill: white;
  }
`;
