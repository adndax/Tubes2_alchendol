"use client" 
import { useSearchParams } from "next/navigation"; 
import { useRouter } from "next/navigation"; 
import { Heading, Paragraph } from "@/components/Typography"; 
import { BorderBox } from "@/components/BorderBox"; 
import { PrimaryButton, SecondaryButton } from "@/components/Button"; 
import { ResultStatCard } from "@/components/Card"; 
import TreeDiagram from "@/components/TreeDiagram"; 
import { useState, useEffect } from "react";  

export default function ResultPage() {   
  const searchParams = useSearchParams();   
  const router = useRouter();   
  const target = searchParams.get("target");   
  const algo = searchParams.get("algo") || "dfs";   
  
  // Read the multiple parameter properly - convert string to boolean   
  const multipleParam = searchParams.get("multiple") || "false";   
  const multiple = multipleParam === "true";   
  
  // Only read maxRecipes if in multiple mode
  const maxRecipes = multiple 
    ? parseInt(searchParams.get("maxRecipes") || "3", 10)
    : 1; // Default to 1 for shortest mode, though it's not used
  
  const [stats, setStats] = useState({      
    nodeCount: 0,      
    timeMs: 0,   
    loading: true,     
    error: null    
  });    
  
  // Debug logging   
  useEffect(() => {     
    console.log("ResultPage params:", {       
      target,       
      algo,       
      multiple,       
      maxRecipes: multiple ? maxRecipes : "N/A" // Show N/A for shortest mode    
    });   
  }, [target, algo, multiple, maxRecipes]);    
  
  if (!target) {     
    return <p className="text-red-500">❌ Target not specified.</p>;   
  }    
  
  // Render loading indicator for stats   
  const renderStatCard = (iconSrc, value, label, isLoading = false) => {     
    return (       
      <ResultStatCard          
        iconSrc={iconSrc}          
        value={isLoading ? "..." : value}          
        label={label}          
        loading={isLoading}        
      />     
    );   
  };    

  // Handle navigation back to element detail page
  const handleBackToElementDetail = () => {
    // Construct URL to go back to the element detail page with current mode
    router.push(`/${multiple ? "multiplerecipes" : "shortestrecipe"}?algo=${algo}`);
  };

  return (     
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">       
      <div className="flex flex-col items-center pt-15 gap-15 w-full pb-20">         
        <div className="flex flex-col gap-4 items-center">           
          <Heading>Eureka! Here's Your Alchemy Route</Heading>           
          <Paragraph>             
            You're searching in <strong>{multiple ? "multiple" : "shortest"}</strong> mode using the <strong>{algo}</strong> spell.           
          </Paragraph>           
          {multiple && (             
            <Paragraph>               
              Showing up to <strong>{maxRecipes}</strong> different recipes.             
            </Paragraph>           
          )}         
        </div>          
        
        <BorderBox className="w-full">           
          <div className="flex flex-col items-center p-10 gap-6">             
            <TreeDiagram                
              target={target}                
              algo={algo}               
              mode={multiple ? "multiple" : "shortest"}               
              quantity={multiple ? maxRecipes : 1}               
              onStatsUpdate={setStats}             
            />              
            
            <div className="flex gap-10 mt-6">               
              {renderStatCard("/img/time.png", `${stats.timeMs}ms`, "Search Time", stats.loading)}               
              {renderStatCard("/img/tree.png", `${stats.nodeCount} nodes`, "Nodes Visited", stats.loading)}               
            </div>                          
            
            {stats.error && !stats.loading && (               
              <div className="mt-4 text-red-500 text-center">                 
                <p className="font-bold">An error occurred:</p>                 
                <p>{stats.error}</p>               
              </div>             
            )}           
          </div>         
        </BorderBox>          
        
        <div className="flex flex-wrap gap-5 justify-center">
          <PrimaryButton 
            label="Try Different Element" 
            onClick={handleBackToElementDetail} 
          />
          <SecondaryButton 
            label="Back To Home" 
            onClick={() => router.push("/")} 
          />
        </div>
      </div>     
    </main>   
  ); 
}

// Tambahkan komponen ini ke ResultPage.js

function ApiDebugConsole({ target, algo, multiple, maxRecipes }) {
  const [apiResponse, setApiResponse] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const checkApi = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Construct API URL
      let url = `http://localhost:8080/api/search?algo=${algo}&target=${encodeURIComponent(target)}`;
      if (multiple) {
        url += `&multiple=true&maxRecipes=${maxRecipes}`;
      }
      
      const response = await fetch(url);
      
      if (!response.ok) {
        throw new Error(`API error: ${response.status} ${response.statusText}`);
      }
      
      const data = await response.json();
      setApiResponse(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mt-6 border border-gray-300 rounded p-4 text-left w-full">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-bold">API Debug Console</h3>
        <button 
          onClick={checkApi}
          className="bg-gray-200 hover:bg-gray-300 text-gray-800 py-1 px-3 rounded"
          disabled={loading}
        >
          {loading ? "Loading..." : "Check API"}
        </button>
      </div>
      
      <div className="mb-2">
        <p><strong>Target:</strong> {target}</p>
        <p><strong>Algorithm:</strong> {algo}</p>
        <p><strong>Multiple:</strong> {multiple ? "Yes" : "No"}</p>
        {multiple && <p><strong>Max Recipes:</strong> {maxRecipes}</p>}
      </div>
      
      {error && (
        <div className="text-red-500 mt-2 mb-2">
          <p><strong>Error:</strong> {error}</p>
        </div>
      )}
      
      {apiResponse && (
        <div className="mt-4">
          <p>
            <strong>Recipe Count:</strong> {
              apiResponse.roots?.length || 
              (Array.isArray(apiResponse.root) ? apiResponse.root.length : 
               (apiResponse.root ? 1 : 0))
            }
          </p>
          <p><strong>Nodes Visited:</strong> {apiResponse.nodesVisited}</p>
          <p><strong>Time Elapsed:</strong> {(apiResponse.timeElapsed * 1000).toFixed(2)}ms</p>
          
          <details>
            <summary className="cursor-pointer mt-2 text-blue-500">View Full Response</summary>
            <pre className="mt-2 bg-gray-100 p-2 rounded text-xs overflow-auto max-h-64">
              {JSON.stringify(apiResponse, null, 2)}
            </pre>
          </details>
        </div>
      )}
    </div>
  );
}

