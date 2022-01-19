// A collection of helper functions to render a graph of kubernetes objects
// with in the context of their parent-child relationships.
import * as d3 from "d3";
import _ from "lodash";
import {
  Application,
  Applications,
  GroupVersionKind,
  UnstructuredObject,
} from "./api/applications/applications.pb";

export type UnstructuredObjectWithParent = UnstructuredObject & {
  parentUid?: string;
};

// Kubernetes does not allow us to query children by parents.
// We keep a list of common parent-child relationships
// to look up children recursively.
export const PARENT_CHILD_LOOKUP = {
  Deployment: {
    group: "apps",
    version: "v1",
    kind: "Deployment",
    children: [
      {
        group: "apps",
        version: "v1",
        kind: "ReplicaSet",
        children: [{ version: "v1", kind: "Pod" }],
      },
    ],
  },
};

export const getChildrenRecursive = async (
  appsClient: typeof Applications,
  result: UnstructuredObjectWithParent[],
  object: UnstructuredObjectWithParent,
  lookup: any
) => {
  result.push(object);

  const k = lookup[object.groupVersionKind.kind];

  if (k && k.children) {
    for (let i = 0; i < k.children.length; i++) {
      const child: GroupVersionKind = k.children[i];

      const res = await appsClient.GetChildObjects({
        parentUid: object.uid,
        groupVersionKind: child,
      });

      for (let q = 0; q < res.objects.length; q++) {
        const c = res.objects[q];

        // Dive down one level and update the lookup accordingly.
        await getChildrenRecursive(
          appsClient,
          result,
          { ...c, parentUid: object.uid },
          {
            [child.kind]: child,
          }
        );
      }
    }
  }
};

// Gets the "child" objects that result from an Application
export const getChildren = async (
  appsClient: typeof Applications,
  app: Application,
  kinds: GroupVersionKind[]
): Promise<UnstructuredObject[]> => {
  const { objects } = await appsClient.GetReconciledObjects({
    automationName: app.name,
    automationNamespace: app.namespace,
    kinds,
  });

  const result = [];
  for (let o = 0; o < objects.length; o++) {
    const obj = objects[o];

    await getChildrenRecursive(appsClient, result, obj, PARENT_CHILD_LOOKUP);
  }

  return _.flatten(result);
};

export const Graph = (opts) => {
  const element = opts.element || {},
    margin = opts.margin || 20,
    width = opts.width || element.offsetWidth || 900,
    height = (opts.height || element.offsetHeight || 600) - 0.5 - margin,
    radius = opts.radius || 5e1,
    distance = opts.distance || 50,
    strength = opts.strength || -8e3,
    graph = opts.graph || {};

  const scale = (opts.dragAndZoom && opts.dragAndZoom.scale) || [0.1, 4];

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

  // ------------------
  // Force Simulation
  // ------------------
  const simulation = d3
    .forceSimulation()
    .force("center", d3.forceCenter(width / 2, height / 2))
    .force("charge", d3.forceManyBody().strength(-50))
    .force("collide", d3.forceCollide().radius(radius + 30));

  simulation.nodes(graph.nodes).on("tick", () => {
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
  simulation
    .force(
      "link",
      d3
        .forceLink()
        .id(function (d) {
          return d.id;
        })
        .iterations(2)
    )
    .force("link")
    .links(graph.links);

  // ----- /Simulation

  //ZOOM
  if (opts.dragAndZoom) {
    const zoom = d3
      .zoom()
      .scaleExtent(scale)
      // .scaleExtent([0.5, 10])
      // .translateExtent([[0, -800], [600, 800]])
      .on("zoom", function zoomed(e) {
        // console.log(e.transform);
        // console.log(e.scale, e.translate) // v3
        g.attr("transform", e.transform);
      });

    svg.call(zoom);

    // reset button
    d3.select(".reset").on("click", function resetted() {
      this.svg.transition().duration(750).call(zoom.transform, d3.zoomIdentity);
    });
  }
};
