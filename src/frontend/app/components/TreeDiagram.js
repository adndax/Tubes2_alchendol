"use client";
import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";
import { elements } from "@data";

export default function TreeDiagram({ target, algo = "DFS", mode = "shortest", maxRecipes = 1, onStatsUpdate }) {
  const ref = useRef();
  const [treeData, setTreeData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [timeoutId, setTimeoutId] = useState(null);
  const [renderAttempted, setRenderAttempted] = useState(false);

  // Create a map for element images - simplified mapping
  const elementImageMap = {};
  elements.forEach(element => {
    elementImageMap[element.name] = element.imageSrc;
  });

  // Effect to handle fetching data
  useEffect(() => {
    if (!target) return;

    // Clear previous timeout if exists
    if (timeoutId) {
      clearTimeout(timeoutId);
    }

    // Reset states on new request
    setLoading(true);
    setError(null);
    setTreeData(null);
    setRenderAttempted(false);

    const formattedAlgo = algo.toUpperCase() === "BIDIRECTIONAL" ? "bidirectional" : algo.toUpperCase();
    
    // Build the correct URL with all parameters
    let url = `http://localhost:8080/api/search?algo=${formattedAlgo}&target=${encodeURIComponent(target)}`;
    
    // Important: Make sure we're passing maxRecipes correctly
    if (mode === "multiple") {
      // Use both mode and multiple parameters to ensure compatibility
      url += `&mode=multiple&multiple=true&maxRecipes=${maxRecipes}`;
    }
    
    console.log("Fetching from URL:", url);
    
    // Set client-side timeout
    const fetchTimeoutId = setTimeout(() => {
      setLoading(false);
      setError(`Request timed out after 15 seconds. The server might be busy or the element "${target}" might be too complex to process.`);
      
      if (onStatsUpdate) {
        onStatsUpdate({ nodeCount: 0, timeMs: 0 });
      }
    }, 15000);
    
    setTimeoutId(fetchTimeoutId);
    
    fetch(url)
      .then((res) => {
        // Clear the timeout when we get a response
        clearTimeout(fetchTimeoutId);
        
        if (!res.ok) {
          if (res.status === 408 || res.status === 504) {
            throw new Error(`Search timeout: The element "${target}" may be too complex to process. Try a different element or reduce maxRecipes.`);
          }
          return res.json().then(errorData => {
            throw new Error(errorData.error || `HTTP error! status: ${res.status}`);
          });
        }
        return res.json();
      })
      .then((data) => {
        console.log("Raw response data:", JSON.stringify(data, null, 2));
        
        // Handle both single and multiple response formats
        let rootData;
        let nodeCount;
        
        if (mode === "multiple") {
          if (data.roots && Array.isArray(data.roots)) {
            if (data.roots.length > 0) {
              // For multiple mode, create a parent node to hold all recipes
              rootData = createMultipleRecipeTree(data.roots, target);
              nodeCount = data.nodesVisited || countNodesInTree(rootData);
            } else {
              // When we have an empty array, show a friendly message
              throw new Error(`No recipes could be found for "${target}" in multiple mode. Try a different element or algorithm.`);
            }
          } else {
            throw new Error('Invalid response format for multiple recipes mode');
          }
        } else {
          // Single recipe mode
          if (data.root) {
            rootData = data.root;
            nodeCount = data.nodesVisited || countNodesInTree(rootData);
          } else {
            throw new Error('No recipe found in response');
          }
        }
        
        if (!rootData) {
          throw new Error('No tree data could be processed');
        }
        
        // Store the tree data in state
        setTreeData(rootData);

        if (onStatsUpdate) {
          const timeMs = Math.round((data.timeElapsed || 0) * 1000);
          onStatsUpdate({ nodeCount, timeMs });
        }

        setLoading(false);
      })
      .catch((err) => {
        console.error("Failed to fetch tree:", err);
        setError(err.message);
        setLoading(false);
        
        if (onStatsUpdate) {
          onStatsUpdate({ nodeCount: 0, timeMs: 0 });
        }
      });
      
    // Cleanup function to clear timeout if component unmounts
    return () => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    };
  }, [target, algo, mode, maxRecipes]);

  // Separate effect for rendering the tree
  useEffect(() => {
    if (ref.current && treeData && !renderAttempted) {
      renderTree(treeData);
      setRenderAttempted(true);
    }
  }, [treeData, renderAttempted]);

  // Add window resize listener to redraw the tree
  useEffect(() => {
    const handleResize = () => {
      if (treeData) {
        renderTree(treeData);
      }
    };

    window.addEventListener('resize', handleResize);
    
    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [treeData]);

  const createMultipleRecipeTree = (roots, targetName) => {
    // Validate all roots have the required structure
    const validRoots = roots.filter(root => {
      return root && (root.Root === targetName || root.root === targetName);
    });
    
    if (validRoots.length === 0) {
      console.warn("No valid roots found in response");
      return null;
    }
    
    // Create a parent node that contains all recipe trees
    return {
      root: targetName,
      Root: targetName,
      isMultipleRoot: true,
      children: validRoots
    };
  };

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

    // Clear previous content
    d3.select(ref.current).selectAll("*").remove();

    console.log("Rendering tree data:", JSON.stringify(data, null, 2));

    const treeDepth = getTreeDepth(data);
    const treeWidth = getTreeWidth(data);
    const circleRadius = 45; // Bigger circle radius like in paste.txt
    
    // Adjust spacing based on mode
    const horizontalSpacing = mode === "multiple" ? 350 : 250;
    const verticalSpacing = 200;
    
    // Calculate dimensions
    const width = Math.max(1600, treeWidth * horizontalSpacing);
    const height = Math.max(800, (treeDepth + 1) * verticalSpacing);
    const margin = { top: 100, right: 150, bottom: 100, left: 150 };

    // Create hierarchy
    const root = d3.hierarchy(data, (d) => d.children);
    
    // Create tree layout
    const treeLayout = d3.tree()
      .size([width - margin.left - margin.right, height - margin.top - margin.bottom])
      .separation((a, b) => {
        // Increase separation for multiple recipes
        if (mode === "multiple" && a.parent === root && b.parent === root) {
          return 3;
        }
        return a.parent === b.parent ? 1 : 2;
      });
    
    // Apply the tree layout
    treeLayout(root);

    // Invert y coordinates (top to bottom)
    root.descendants().forEach((d) => {
      d.y = (height - margin.top - margin.bottom) - d.y;
    });

    // Create SVG
    const svg = d3.select(ref.current)
      .attr("width", width)
      .attr("height", height)
      .attr("viewBox", `0 0 ${width} ${height}`)
      .attr("style", "max-width: 100%; height: auto;");

    // Background
    svg.append("rect")
      .attr("width", width)
      .attr("height", height)
      .attr("fill", "#FFE1A8");

    // Create main group
    const g = svg.append("g")
      .attr("transform", `translate(${margin.left},${margin.top})`);

    // Define clip paths for circular images
    const defs = svg.append("defs");
    
    root.descendants().forEach((d, i) => {
      defs.append("clipPath")
        .attr("id", `clip-${i}`)
        .append("circle")
        .attr("cx", 0)
        .attr("cy", 0)
        .attr("r", 30); // Image clip radius
    });

    // Color palette for different recipes
    const colors = ["#8B4513", "#D2691E", "#CD853F", "#A0522D", "#B8860B", "#654321", "#8B6914", "#A0522D"];
    
    // Draw elbow-style links with error handling
    try {
      g.selectAll("path.link")
        .data(root.links())
        .join("path")
        .attr("class", "link")
        .attr("d", (d) => {
          if (!d.source || !d.target) return "";
          
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
        .attr("stroke", d => {
          // Color code different recipe branches
          if (mode === "multiple" && d.source.data.isMultipleRoot && d.target.parent === root) {
            const childIndex = root.children.indexOf(d.target);
            return colors[childIndex % colors.length];
          }
          // Maintain color throughout the branch
          else if (mode === "multiple" && d.target.depth > 1) {
            let currentNode = d.target;
            while (currentNode.parent && currentNode.parent !== root) {
              currentNode = currentNode.parent;
            }
            if (currentNode.parent === root) {
              const branchIndex = root.children.indexOf(currentNode);
              return colors[branchIndex % colors.length];
            }
          }
          return "#666";
        })
        .attr("stroke-width", d => {
          // Thicker lines for main recipe branches
          return (mode === "multiple" && d.source.data.isMultipleRoot) ? 4 : 2;
        });
      
      // Draw nodes
      const nodes = g.selectAll("g.node")
        .data(root.descendants())
        .enter()
        .append("g")
        .attr("class", "node")
        .attr("transform", d => `translate(${d.x},${d.y})`);

      // Add circles for nodes (bigger like in paste.txt)
      nodes.append("circle")
        .attr("r", circleRadius)
        .attr("fill", d => {
          const elementName = d.data.root || d.data.Root || d.data.name || d.data.element || "";
          const basicElements = ["Air", "Earth", "Fire", "Water"];
          
          // Special color for multiple root node
          if (mode === "multiple" && d.data.isMultipleRoot) {
            return "#4B0082"; // Indigo for the main target
          }
          return basicElements.includes(elementName) ? "#6B8E23" : "#A6352B";
        })
        .attr("stroke", "#333")
        .attr("stroke-width", 2);

      // Add images inside circles (positioned in upper part)
      nodes.each(function(d, i) {
        const elementName = d.data.root || d.data.Root || d.data.name || d.data.element || "";
        const imageSrc = elementImageMap[elementName];
        const node = d3.select(this);
        
        if (imageSrc) {
          const image = node.append("image")
            .attr("xlink:href", imageSrc)
            .attr("x", -22.5) // Center the image
            .attr("y", -30) // Position in upper part of circle
            .attr("width", 45)
            .attr("height", 45)
            .attr("clip-path", `url(#clip-${i})`)
            .on("error", function() {
              // On error, remove image and fall back to text only
              d3.select(this).remove();
            });
        }
      });

      // Add text inside circle (positioned in lower part)
      nodes.append("text")
        .attr("y", 20) // Position in lower part of circle
        .attr("text-anchor", "middle")
        .attr("dominant-baseline", "middle")
        .text((d) => {
          const name = d.data.root || d.data.Root || d.data.name || d.data.element || "";
          // Truncate long names
          return name.length > 10 ? name.substring(0, 9) + "..." : name;
        })
        .attr("font-size", 14)
        .attr("fill", "white") // White text to contrast with dark background
        .attr("font-weight", "bold");

      // Add recipe labels for multiple mode
      if (mode === "multiple" && data.isMultipleRoot) {
        nodes.filter(d => d.parent === root && d !== root)
          .append("text")
          .attr("text-anchor", "middle")
          .attr("y", -circleRadius - 20)
          .text((d, i) => `Recipe ${i + 1}`)
          .attr("font-size", 16)
          .attr("fill", d => {
            const index = root.children.indexOf(d);
            return colors[index % colors.length];
          })
          .attr("font-weight", "bold");
      }

      // Add main title for multiple recipes
      if (mode === "multiple") {
        svg.append("text")
          .attr("x", width / 2)
          .attr("y", 40)
          .attr("text-anchor", "middle")
          .attr("font-size", 24)
          .attr("font-weight", "bold")
          .attr("fill", "#333")
          .text(`All Recipes for ${target}`);
      }
    } catch (renderError) {
      console.error("Error rendering tree:", renderError);
      
      // Display error message on the SVG
      svg.append("text")
        .attr("x", width / 2)
        .attr("y", height / 2)
        .attr("text-anchor", "middle")
        .attr("font-size", 18)
        .attr("fill", "red")
        .text(`Error rendering tree: ${renderError.message}`);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-lg">Loading tree diagram...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col justify-center items-center h-64">
        <div className="text-red-500 text-lg">Error: {error}</div>
        <div className="mt-4 text-sm">
          <p>Make sure the backend server is running:</p>
          <code className="bg-gray-100 p-2 rounded">cd backend && go run main.go</code>
          <p className="mt-2">Try a less complex element or reduce maxRecipes if the search times out.</p>
        </div>
      </div>
    );
  }

  return <svg ref={ref}></svg>;
}