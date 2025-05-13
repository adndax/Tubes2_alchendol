"use client";
import { useEffect, useRef, useState } from "react";
import * as d3 from "d3";
import { elements } from "@data";

export default function TreeDiagram({ target, algo = "DFS", mode = "shortest", maxRecipes = 5, onStatsUpdate }) {
  const ref = useRef();
  const [treeData, setTreeData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [timeoutId, setTimeoutId] = useState(null);
  const [renderAttempted, setRenderAttempted] = useState(false);
  // Tambahkan state untuk navigasi multiple recipe
  const [recipes, setRecipes] = useState([]);
  const [currentRecipeIndex, setCurrentRecipeIndex] = useState(0);
  const [totalRecipes, setTotalRecipes] = useState(0);

  // Create a map for element images - simplified mapping
  const elementImageMap = {};
  elements.forEach(element => {
    elementImageMap[element.name] = element.imageSrc;
  });

  // List of basic elements and special elements
  const basicElements = ["Air", "Earth", "Fire", "Water"];
  const specialElements = ["Clock", "Death", "Dinosaur", "Family Tree", "Peat", "Skeleton", "Sloth", "Tree"];

  // Determine if element is complex and needs longer timeout
  const isComplexElement = (elementName) => {
    const complexElements = ["Picnic", "Skyscraper", "City", "Continent", "Horseshoe", "Unicorn"];
    return complexElements.includes(elementName);
  };

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
    setRecipes([]);
    setCurrentRecipeIndex(0);
    setTotalRecipes(0);

    const formattedAlgo = algo.toUpperCase() === "BIDIRECTIONAL" ? "bidirectional" : algo.toUpperCase();
    
    // Build the correct URL with all parameters
    let url = `http://localhost:8080/api/search?algo=${formattedAlgo}&target=${encodeURIComponent(target)}`;
    
    // Important: Make sure we're passing maxRecipes correctly for multiple mode
    if (mode === "multiple") {
      // Use both mode and multiple parameters to ensure compatibility
      url += `&mode=multiple&multiple=true&maxRecipes=${maxRecipes}`;
    }
    
    console.log("Fetching from URL:", url);
    
    // Set longer timeout for complex elements
    const timeoutDuration = isComplexElement(target) ? 30000 : 15000; // 30 seconds for complex elements
    
    // Set client-side timeout
    const fetchTimeoutId = setTimeout(() => {
      setLoading(false);
      setError(`Request timed out after ${timeoutDuration/1000} seconds. The server might be busy or the element "${target}" might be too complex to process.`);
      
      if (onStatsUpdate) {
        onStatsUpdate({ nodeCount: 0, timeMs: 0 });
      }
    }, timeoutDuration);
    
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
              // Untuk multiple mode, simpan semua resep secara terpisah
              setRecipes(data.roots);
              setCurrentRecipeIndex(0);
              setTotalRecipes(data.roots.length);
              
              // Pilih resep pertama untuk ditampilkan
              rootData = data.roots[0];
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
          const timeMs = data.timeElapsed || 0;
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

  // Effect untuk navigasi recipe
  useEffect(() => {
    if (mode === "multiple" && recipes.length > 0) {
      setTreeData(recipes[currentRecipeIndex]);
      setRenderAttempted(false);
    }
  }, [currentRecipeIndex, recipes, mode]);

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

  const countNodesInTree = (node) => {
    if (!node) return 0;
    let count = 1;
    const children = node.children || node.Children || [];
    
    for (const child of children) {
      count += countNodesInTree(child);
    }
    
    return count;
  };

  const getTreeDepth = (node, depth = 0) => {
    if (!node) return depth;
    
    const children = node.children || node.Children || [];
    
    if (children.length === 0) return depth;
    
    return Math.max(...children.map((child) => getTreeDepth(child, depth + 1)));
  };

  const getTreeWidth = (node) => {
    if (!node) return 0;
    
    const widthByLevel = {};
    
    const countByLevel = (node, level = 0) => {
      widthByLevel[level] = (widthByLevel[level] || 0) + 1;
      
      const children = node.children || node.Children || [];
      children.forEach((child) => countByLevel(child, level + 1));
    };
    
    countByLevel(node);
    
    return Math.max(...Object.values(widthByLevel), 0);
  };

  // Normalize the tree data to handle different property naming conventions
  const normalizeTree = (node) => {
    if (!node) return null;
    
    // Create a new normalized node
    const normalized = {
      name: node.root || node.Root || node.element || node.name || "",
      children: []
    };
    
    // Normalize children
    const sourceChildren = node.children || node.Children || [];
    normalized.children = sourceChildren.map(child => normalizeTree(child));
    
    return normalized;
  };

  const renderTree = (data) => {
    if (!ref.current || !data) return;

    // Clear previous content
    d3.select(ref.current).selectAll("*").remove();

    console.log("Rendering tree data:", JSON.stringify(data, null, 2));
    
    try {
      // Normalize the tree structure for consistent property access
      const normalizedData = normalizeTree(data);
      
      if (!normalizedData) {
        console.error("Failed to normalize tree data");
        return;
      }
      
      const treeDepth = getTreeDepth(data);
      const treeWidth = getTreeWidth(data);
      const circleRadius = 45; // Circle radius
      
      // Adjust spacing - lebih besar untuk pembacaan yang lebih mudah
      const horizontalSpacing = 250;
      const verticalSpacing = 200;
      
      // Calculate dimensions
      const width = Math.max(1200, treeWidth * horizontalSpacing);
      const height = Math.max(800, (treeDepth + 1) * verticalSpacing);
      const margin = { top: 100, right: 150, bottom: 100, left: 150 };

      // Create hierarchy with the normalized data
      const root = d3.hierarchy(normalizedData, d => d.children);
      
      // Create tree layout
      const treeLayout = d3.tree()
        .size([width - margin.left - margin.right, height - margin.top - margin.bottom])
        .separation((a, b) => {
          return a.parent === b.parent ? 1.5 : 2.5; // Increased separation for better readability
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

      // Color for the current recipe
      const mainColor = "#8B4513";
      
      // Draw links - with error protection
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
        .attr("stroke", mainColor)
        .attr("stroke-width", 3);
      
      // Draw nodes
      const nodes = g.selectAll("g.node")
        .data(root.descendants())
        .enter()
        .append("g")
        .attr("class", "node")
        .attr("transform", d => `translate(${d.x},${d.y})`);

      // Add circles for nodes with proper coloring
      nodes.append("circle")
        .attr("r", circleRadius)
        .attr("fill", d => {
          const elementName = d.data.name;
          
          // Check if this is one of the special elements that should be purple
          if (specialElements.includes(elementName)) {
            return "#800080"; // Purple for special elements
          }
          
          // Original colors for other elements
          if (basicElements.includes(elementName)) {
            return "#6B8E23"; // Green for basic elements
          }
          
          return "#A6352B"; // Red for other elements
        })

      // Add images inside circles
      nodes.each(function(d, i) {
        const elementName = d.data.name;
        const imageSrc = elementImageMap[elementName];
        const node = d3.select(this);
        
        if (imageSrc) {
          node.append("image")
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

      // Add text inside circle
      nodes.append("text")
        .attr("y", 20) // Position in lower part of circle
        .attr("text-anchor", "middle")
        .attr("dominant-baseline", "middle")
        .text((d) => {
          const name = d.data.name;
          // Truncate long names
          return name.length > 10 ? name.substring(0, 9) + "..." : name;
        })
        .attr("font-size", 14)
        .attr("fill", "white")
        .attr("font-weight", "bold");

      // Tambahkan judul
      if (mode === "multiple" && totalRecipes > 0) {
        svg.append("text")
          .attr("x", width / 2)
          .attr("y", 40)
          .attr("text-anchor", "middle")
          .attr("font-size", 24)
          .attr("font-weight", "bold")
          .attr("fill", "#333")
          .text(`Recipe ${currentRecipeIndex + 1} of ${totalRecipes} for ${target}`);
      } else {
        svg.append("text")
          .attr("x", width / 2)
          .attr("y", 40)
          .attr("text-anchor", "middle")
          .attr("font-size", 24)
          .attr("font-weight", "bold")
          .attr("fill", "#333")
          .text(`Recipe for ${target}`);
      }
    } catch (err) {
      console.error("Error rendering tree:", err);
      setError(`Error rendering tree: ${err.message}`);
    }
  };

  // Handler untuk navigasi
  const handlePrevRecipe = () => {
    if (currentRecipeIndex > 0) {
      setCurrentRecipeIndex(currentRecipeIndex - 1);
    }
  };

  const handleNextRecipe = () => {
    if (currentRecipeIndex < totalRecipes - 1) {
      setCurrentRecipeIndex(currentRecipeIndex + 1);
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
      <div className="flex flex-col justify-center items-center h-64 p-4 rounded-lg shadow-md" style={{ backgroundColor: '#2A0026' }}>
        <h2 className="text-xl font-semibold mb-2" style={{ color: '#FFE1A8' }}>An error occurred</h2>
        <p className="mb-4" style={{ color: '#ffffff' }}>
          Details: <span style={{ color: '#A6352B' }}>{error}</span>
        </p>
        <div className="p-3 rounded-md w-full" style={{ backgroundColor: '#7d2820' }}>
          <p className="mb-2" style={{ color: '#FFE1A8' }}>Ensure the backend server is running:</p>
          <code className="block p-2 rounded-md" style={{ backgroundColor: '#A6352B', color: '#ffffff' }}>
            cd backend && go run main.go
          </code>
        </div>
        <p className="text-s mt-3" style={{ color: '#FFE1A8' }}>
          If the search times out, try a simpler query or reduce <code style={{ color: '#A6352B' }}>maxRecipes</code>.
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center">
      {mode === "multiple" && totalRecipes > 0 && (
        <div className="flex items-center justify-center mb-4 space-x-4">
          <button 
            onClick={handlePrevRecipe}
            disabled={currentRecipeIndex === 0}
            className={`px-4 py-2 rounded-lg font-semibold ${
              currentRecipeIndex === 0 
                ? 'bg-gray-300 text-gray-500 cursor-not-allowed border-2 border-gray-500' 
                : 'bg-primary text-white hover:bg-primary-hover border-2 border-secondary'
            }`}
          >
            Previous Recipe
          </button>
          <span className="text-lg font-semibold">
            {currentRecipeIndex + 1} / {totalRecipes}
          </span>
          <button 
            onClick={handleNextRecipe}
            disabled={currentRecipeIndex === totalRecipes - 1}
            className={`px-4 py-2 rounded-lg font-semibold ${
              currentRecipeIndex === totalRecipes - 1 
                ? 'bg-gray-300 text-gray-500 cursor-not-allowed border-2 border-gray-500' 
                : 'bg-primary text-white hover:bg-primary-hover border-2 border-secondary'
            }`}
          >
            Next Recipe
          </button>
        </div>
      )}
      <svg ref={ref}></svg>
    </div>
  );
}