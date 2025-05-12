"use client";
import { SecondaryButton } from "@/components/Button";
import { Heading, Paragraph } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";
import Image from "next/image";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";

export default function SearchingPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  // Get search parameters
  const target = searchParams.get("target");
  const algo = searchParams.get("algo") || "dfs";
  const multipleParam = searchParams.get("multiple") || "false";
  const multiple = multipleParam === "true";
  const maxRecipes = parseInt(searchParams.get("maxRecipes") || "3", 10);
  
  // Tracking the search request to allow cancellation
  const [abortController, setAbortController] = useState(null);
  const [isSearchCancelled, setIsSearchCancelled] = useState(false);
  
  // Perform search when component mounts
  useEffect(() => {
    // If no target, go home
    if (!target) {
      router.push("/");
      return;
    }
    
    // If search was cancelled, don't continue
    if (isSearchCancelled) return;
    
    // Create abort controller for cancellation
    const controller = new AbortController();
    setAbortController(controller);
    
    // Construct API URL
    const apiUrl = multiple 
      ? `http://localhost:8080/api/search?algo=${algo}&target=${target}&multiple=true&maxRecipes=${maxRecipes}`
      : `http://localhost:8080/api/search?algo=${algo}&target=${target}`;
    
    console.log("Fetching from API:", apiUrl);
    
    // Fetch data with abort signal
    fetch(apiUrl, { signal: controller.signal })
      .then((res) => {
        if (!res.ok) {
          throw new Error(`API responded with status ${res.status}`);
        }
        return res.json();
      })
      .then((data) => {
        // Log response for debugging
        console.log("API response:", JSON.stringify(data, null, 2));
        
        // Only navigate if search wasn't cancelled
        if (!isSearchCancelled) {
          // Navigate to result page with appropriate parameters
          // Don't include maxRecipes for shortest mode
          let resultUrl = `/result?target=${encodeURIComponent(target)}&algo=${algo}&multiple=${multiple}`;
          if (multiple) {
            resultUrl += `&maxRecipes=${maxRecipes}`;
          }
          router.push(resultUrl);
        }
      })
      .catch((err) => {
        // If this is an abort error, ignore it (user cancelled)
        if (err.name === 'AbortError') {
          console.log('Search was cancelled');
          return;
        }
        
        console.error("Failed to fetch tree:", err);
        // Navigate to element page with error
        const errorMsg = encodeURIComponent(err.message || "Failed to fetch data");
        router.push(`/element/${target}?mode=${multiple ? "multiple" : "shortest"}&algo=${algo}&error=${errorMsg}`);
      });
    
    // Cleanup function
    return () => {
      controller.abort();
    };
  }, [router, target, algo, multiple, maxRecipes, isSearchCancelled]);
  
  // Handle cancel button click
  const handleCancel = () => {
    setIsSearchCancelled(true);
    if (abortController) {
      abortController.abort();
    }
    
    // Navigate back to the element detail page
    if (target) {
      router.push(`/element/${target}?mode=${multiple ? "multiple" : "shortest"}&algo=${algo}`);
    } else {
      router.push('/');
    }
  };
  
  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-28 gap-22 w-full pb-20">
        <div className="flex flex-col gap-8 items-center">
          <Heading>Brewing the Recipe…</Heading>
          <Paragraph>
            Meowculus is sniffing through the archives…
          </Paragraph>
        </div>

        <BorderBox className="w-full">
          <div className="flex flex-col items-center p-10 gap-6">
            <div className="relative">
              <Image
                src="/img/meowculus_stir.png"
                alt="Meowculus"
                width={110}
                height={110}
                className="animate-spin-slow"
              />
            </div>
            <Paragraph>
              {target ? `Searching for ${target} ${multiple ? 'recipes' : 'recipe'}...` : 'Searching...'}
            </Paragraph>
            <Paragraph>
              {multiple ? `Up to ${maxRecipes} recipes will be found` : 'Finding the shortest recipe'}
            </Paragraph>
          </div>
        </BorderBox>

        <SecondaryButton onClick={handleCancel} label="Cancel"/>
      </div>
    </main>
  );
}