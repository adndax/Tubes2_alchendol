"use client";
import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";
import { elements } from "@data";

export default function TreeDiagram({ target, algo = "dfs", onStatsUpdate }) {
  const ref = useRef();
  const [treeData, setTreeData] = useState(null);

  // Buat mapping dari nama element ke imageSrc
  const elementImageMap = {};
  elements.forEach(element => {
    elementImageMap[element.name] = element.imageSrc;
  });

  useEffect(() => {
    if (!target) return;

    fetch(`http://localhost:8080/api/search?algo=${algo}&target=${target}`)
      .then((res) => res.json())
      .then((data) => {
        const rootData = data.root || data;
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
    const width = Math.max(800, treeWidth * 250); // Increased width for bigger circles
    const height = Math.max(500, treeDepth * 200); // Increased height
    const margin = { top: 100, right: 100, bottom: 100, left: 100 };

    const root = d3.hierarchy(data, (d) => d.children);
    const treeLayout = d3.tree().size([width - margin.left - margin.right, height - margin.top - margin.bottom]);
    treeLayout(root);
    
    // Invert the y-coordinates to have root at top
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

    // Define clip path for circular images
    const defs = svg.append("defs");

    // Create clip paths for each node
    root.descendants().forEach((d, i) => {
      defs.append("clipPath")
        .attr("id", `clip-${i}`)
        .append("circle")
        .attr("cx", 0)
        .attr("cy", 0)
        .attr("r", 30); // Image clip radius
    });

    // Define offsets
    const circleRadius = 45; // Much bigger circle radius

    // Create elbow paths
    g.selectAll("path")
      .data(root.links())
      .join("path")
      .attr("d", (d) => {
        const sourceY = d.source.y + circleRadius; // Start from bottom of circle
        const targetY = d.target.y - circleRadius; // End at top of circle
        
        const midY = (sourceY + targetY) / 2;
        
        // Create elbow path
        return `M ${d.source.x} ${sourceY}
                L ${d.source.x} ${midY}
                L ${d.target.x} ${midY}
                L ${d.target.x} ${targetY}`;
      })
      .attr("fill", "none")
      .attr("stroke", "#666")
      .attr("stroke-width", 2);

    // Create groups for each node
    const nodeGroups = g.selectAll("g.node")
      .data(root.descendants())
      .join("g")
      .attr("class", "node")
      .attr("transform", (d) => `translate(${d.x},${d.y})`);

    // Add circle background (bigger)
    nodeGroups.append("circle")
      .attr("r", circleRadius)
      .attr("fill", d => {
        const elementName = d.data.root || d.data.name || d.data.element || "";
        const basicElements = ["Air", "Earth", "Fire", "Water"];
        return basicElements.includes(elementName) ? "#6B8E23" : "#A6352B";
      }
    )
      .attr("stroke-width", 2);

    // Add images inside circles (positioned in upper part)
    nodeGroups.append("image")
      .attr("xlink:href", (d) => {
        const elementName = d.data.root || d.data.name || d.data.element || "";
        const imageSrc = elementImageMap[elementName];
        return imageSrc || "/icons/placeholder.svg";
      })
      .attr("x", -25) // Center the image
      .attr("y", -30) // Position in upper part of circle
      .attr("width", 45)
      .attr("height", 45)
      .attr("clip-path", (d, i) => `url(#clip-${i})`)
      .on("error", function() {
        d3.select(this).attr("xlink:href", "/icons/placeholder.svg");
      });

    // Add text inside circle (positioned in lower part)
    nodeGroups.append("text")
      .attr("y", 27) // Position in lower part of circle
      .attr("text-anchor", "middle")
      .attr("dominant-baseline", "middle")
      .text((d) => d.data.root || d.data.name || d.data.element || "")
      .attr("font-size", 14)
      .attr("fill", "white") // Changed to white to contrast with dark background
      .attr("font-weight", "bold");
  };

  return <svg ref={ref}></svg>;
}