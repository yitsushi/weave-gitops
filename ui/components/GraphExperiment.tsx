import * as React from "react";
import styled from "styled-components";
import { Graph } from "../lib/graph";

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

  const divRef = React.useRef();

  React.useEffect(() => {
    if (!divRef.current) return;
    else {
      const graph = new Graph({
        element: divRef.current,
        width: 1100,
        height: 600,
        strength: -50,
        radius: 50,
        graph: data,
        dragAndZoom: { scale: [0.1, 2] },
      });
    }
  }, [divRef]);
  return <div className={className} ref={divRef} id="graph"></div>;
}

export default styled(GraphExperiment)`
  .links line {
    stroke-width: 2;
    stroke: white;
    stroke: ${(props) => props.theme.colors.primary};
  }

  .nodes circle {
    stroke-width: 2;
    stroke: white;
    stroke: ${(props) => props.theme.colors.primary};
    fill: white;
  }
`;
