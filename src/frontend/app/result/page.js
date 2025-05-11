"use client"
import { useSearchParams } from "next/navigation";
import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { Heading, Paragraph } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";
import { PrimaryButton, SecondaryButton } from "@/components/Button";
import { ResultStatCard } from "@/components/Card";
import TreeDiagram from "@/components/TreeDiagram";

export default function ResultPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  
  // Get search parameters from URL
  const target = searchParams.get("target");
  const algo = searchParams.get("algo") || "DFS";
  const mode = searchParams.get("mode") || "shortest";
  const quantity = parseInt(searchParams.get("quantity") || "5", 10);
  
  // Store stats and parameters in state to persist through refreshes
  const [stats, setStats] = useState({ nodeCount: 0, timeMs: 0 });
  const [error, setError] = useState(null);
  const [savedParams, setSavedParams] = useState({
    target: null,
    algo: null,
    mode: null,
    quantity: null
  });

  // Save parameters to localStorage when they change
  useEffect(() => {
    if (target && algo && mode) {
      // Store params in state
      setSavedParams({
        target,
        algo,
        mode,
        quantity
      });
      
      // Store params in localStorage
      localStorage.setItem('searchParams', JSON.stringify({
        target,
        algo,
        mode,
        quantity
      }));
    }
  }, [target, algo, mode, quantity]);

  // Restore parameters from localStorage on page load
  useEffect(() => {
    // Only restore from localStorage if params are missing from URL
    if (!target) {
      try {
        const savedParamsJson = localStorage.getItem('searchParams');
        if (savedParamsJson) {
          const savedParams = JSON.parse(savedParamsJson);
          
          // If we have saved params and no current params, redirect to keep params in URL
          if (savedParams.target && !target) {
            let url = `/result?target=${encodeURIComponent(savedParams.target)}&algo=${savedParams.algo || 'DFS'}`;
            if (savedParams.mode) {
              url += `&mode=${savedParams.mode}`;
            }
            if (savedParams.quantity) {
              url += `&quantity=${savedParams.quantity}`;
            }
            
            router.replace(url);
          }
        }
      } catch (err) {
        console.error("Error restoring parameters:", err);
      }
    }
  }, [target, router]);

  // Validate the input parameters
  useEffect(() => {
    if (!target && !savedParams.target) {
      setError("Target element not specified. Please go back and select an element.");
      return;
    }
    
    const currentAlgo = algo || savedParams.algo || "DFS";
    // Case-insensitive validation of algorithm
    const validAlgos = ["DFS", "BFS", "BIDIRECTIONAL"];
    if (!validAlgos.some(validAlgo => validAlgo.toLowerCase() === currentAlgo.toLowerCase())) {
      setError(`Unknown algorithm: ${currentAlgo}. Please use DFS, BFS, or BIDIRECTIONAL.`);
      return;
    }
    
    const currentMode = mode || savedParams.mode;
    const currentQuantity = quantity || savedParams.quantity;
    if (currentMode === "multiple" && (isNaN(currentQuantity) || currentQuantity < 1)) {
      setError("Invalid quantity for multiple recipes. Please specify a positive number.");
      return;
    }
    
    // Clear error if validation passes
    setError(null);
  }, [target, algo, mode, quantity, savedParams]);

  const handleTryAgain = () => {
    router.push("/");
  };

  // Get the effective parameters to use
  const effectiveTarget = target || savedParams.target;
  // Normalize algorithm to uppercase for consistency
  const effectiveAlgo = (algo || savedParams.algo || "DFS").toUpperCase();
  const effectiveMode = mode || savedParams.mode || "shortest";
  const effectiveQuantity = quantity || savedParams.quantity || 5;

  if (error) {
    return (
      <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
        <div className="flex flex-col items-center pt-15 gap-15 w-full pb-20">
          <BorderBox className="w-full">
            <div className="flex flex-col items-center p-10 gap-6">
              <div className="text-red-500 text-lg font-bold mb-4">Error</div>
              <div className="text-primary">{error}</div>
              <div className="mt-6">
                <PrimaryButton label="Go Back to Home" onClick={handleTryAgain} />
              </div>
            </div>
          </BorderBox>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-15 gap-15 w-full pb-20">
        <div className="flex flex-col gap-4 items-center">
          <Heading>Eureka! Here's Your Alchemy Route</Heading>
          <Paragraph>
            You searched, I conjured, and here it is — your magical recipe revealed!
          </Paragraph>
        </div>

        <BorderBox className="w-full">
          <div className="flex flex-col items-center p-10 gap-6">
            <TreeDiagram 
              target={effectiveTarget} 
              algo={effectiveAlgo}
              mode={effectiveMode}
              maxRecipes={effectiveQuantity}
              onStatsUpdate={({ nodeCount, timeMs }) => {
                setStats({ nodeCount, timeMs });
                
                // Also save stats in localStorage
                localStorage.setItem('searchStats', JSON.stringify({ 
                  nodeCount, 
                  timeMs 
                }));
              }}
            />

            <div className="flex gap-10 mt-6">
              <ResultStatCard 
                iconSrc="/img/time.png" 
                value={`${stats.timeMs}ms`} 
                label="Search Time" 
              />
              <ResultStatCard 
                iconSrc="/img/tree.png" 
                value={`${stats.nodeCount} nodes`} 
                label="Nodes Visited" 
              />
            </div>
          </div>
        </BorderBox>

        <div className="flex gap-4 mt-4">
          <PrimaryButton label="Back To Home" onClick={() => router.push("/")} />
          <SecondaryButton 
            label="Search Another Element" 
            onClick={() => router.push(effectiveMode === "multiple" ? "/multiplerecipes" : "/shortestrecipe")} 
          />
        </div>
      </div>
    </main>
  );
}