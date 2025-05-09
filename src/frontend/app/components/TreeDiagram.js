"use client";
import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";

export default function TreeDiagram({ target, algo = "DFS", onStatsUpdate }) {
  const ref = useRef();
  const [treeData, setTreeData] = useState(null);

  useEffect(() => {
    if (!target) return;

    fetch(`http://localhost:8080/api/search?algo=${algo}&target=${target}`)
      .then((res) => res.json())
      .then((data) => {
        const rootData = data.root || data; // if wrapped in {root: {...}}
        setTreeData(rootData);

        if (onStatsUpdate) {
          const nodeCount = countNodesInTree(rootData);
          const timeMs = Math.round((data.timeElapsed || 0) * 1000);
          onStatsUpdate({ nodeCount, timeMs });
        }

        renderTree(rootData);
      })
      .catch((err) => {
        console.error("Failed to fetch tree:", err);
      });
  }, [target, algo]);

  const countNodesInTree = (node) => {
    if (!node) return 0;
    let count = 1;
    if (node.children && node.children.length > 0) {
      node.children.forEach((child) => {
        count += countNodesInTree(child);
      });
    }
    return count;
  };

  const getTreeDepth = (node, depth = 0) => {
    if (!node.children || node.children.length === 0) return depth;
    return Math.max(...node.children.map((child) => getTreeDepth(child, depth + 1)));
  };

  const getTreeWidth = (node) => {
    const widthByLevel = {};
    const countByLevel = (node, level = 0) => {
      widthByLevel[level] = (widthByLevel[level] || 0) + 1;
      node.children?.forEach((child) => countByLevel(child, level + 1));
    };
    countByLevel(node);
    return Math.max(...Object.values(widthByLevel));
  };

  const renderTree = (data) => {
    if (!ref.current || !data) return;

    d3.select(ref.current).selectAll("*").remove();

    const treeDepth = getTreeDepth(data);
    const treeWidth = getTreeWidth(data);
    const width = Math.max(800, treeWidth * 220);
    const height = Math.max(600, treeDepth * 150);
    const margin = { top: 100, right: 100, bottom: 100, left: 100 };

    const root = d3.hierarchy(data, (d) => d.children);
    const treeLayout = d3.tree().size([width - margin.left - margin.right, height - margin.top - margin.bottom]);
    treeLayout(root);

    root.descendants().forEach((d) => {
      d.y = (height - margin.top - margin.bottom) - d.y;
    });

    const svg = d3.select(ref.current)
      .attr("width", width)
      .attr("height", height)
      .attr("viewBox", `0 0 ${width} ${height}`)
      .attr("style", "max-width: 100%; height: auto;");

    svg.append("rect").attr("width", width).attr("height", height).attr("fill", "#FFE1A8");

    const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);

    g.selectAll("line")
      .data(root.links())
      .join("line")
      .attr("x1", (d) => d.source.x)
      .attr("y1", (d) => d.source.y)
      .attr("x2", (d) => d.target.x)
      .attr("y2", (d) => d.target.y)
      .attr("stroke", "#666")
      .attr("stroke-width", 2);

    g.selectAll("circle")
      .data(root.descendants())
      .join("circle")
      .attr("cx", (d) => d.x)
      .attr("cy", (d) => d.y)
      .attr("r", 40)
      .attr("fill", "#A6352B");

    g.selectAll("text")
      .data(root.descendants())
      .join("text")
      .attr("x", (d) => d.x)
      .attr("y", (d) => d.y)
      .attr("text-anchor", "middle")
      .attr("dominant-baseline", "middle")
      .text((d) => d.data.root || d.data.name || d.data.element || "")
      .attr("font-size", 12)
      .attr("fill", "white");
  };

  return <svg ref={ref}></svg>;
}