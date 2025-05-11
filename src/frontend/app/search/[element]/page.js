// In page.js for the element detail page
"use client";
import { useParams, useSearchParams, useRouter } from "next/navigation";
import { useState } from "react";
import { elements } from "@data";
import { Heading, Paragraph, Subheading } from "@/components/Typography";
import { ElementsCard } from "@/components/Card";
import { BorderBox } from "@/components/BorderBox";
import { PrimaryButton } from "@/components/Button";
import QuantityInput from "@/components/QuantityInput";

export default function ElementDetailPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();
  
  const element = decodeURIComponent(params.element);
  const mode = searchParams.get("mode") || "shortest";
  const algo = searchParams.get("algo") || "DFS";
  
  // Default to 5 recipes for multiple mode
  const [quantity, setQuantity] = useState(5);
  const data = elements.find((el) => el.name === element);

  if (!data) {
    return <p className="text-center text-red-500">Element not found</p>;
  }

  const handleSearch = () => {
    // Store the search parameters in localStorage for persistence
    localStorage.setItem('searchParams', JSON.stringify({
      target: element,
      algo: algo,
      mode: mode,
      quantity: quantity
    }));
    
    // Build the search URL with all parameters properly encoded
    let searchUrl = `/result?target=${encodeURIComponent(element)}&algo=${encodeURIComponent(algo)}`;
    
    // Add mode and quantity for multiple mode
    if (mode === "multiple") {
      searchUrl += `&mode=${encodeURIComponent(mode)}&quantity=${quantity}`;
    }
    
    // Debug the URL before navigation
    console.log("Navigating to:", searchUrl);
    
    router.push(searchUrl);
  };

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-28 gap-22 w-full pb-20">
        <div className="flex flex-col gap-2 items-center">
          <Heading>The Final Ingredient... Is You!</Heading>
          <Paragraph>
            The elements are ready… the scrolls have aligned… now all we need is a tiny click!
          </Paragraph>
        </div>

        <BorderBox className="w-full">
          <div className="flex flex-col items-center p-10 gap-6">
            <ElementsCard name={data.name} imageSrc={data.imageSrc} />
            <Paragraph>{data.description || `Create a ${data.name} with the magic of alchemy!`}</Paragraph>
            
            {mode === "multiple" && (
              <>
                <div className="flex flex-col items-center gap-2">
                  <Subheading>How many recipes do you want to discover?</Subheading>
                  <div className="flex items-center gap-2">
                    <QuantityInput 
                      value={quantity} 
                      onChange={(val) => {
                        // Cap the max recipes at 20 to prevent browser slowdowns
                        setQuantity(Math.min(val, 20));
                      }} 
                    />
                    {quantity > 10 && (
                      <Paragraph className="text-red-500 text-xs ml-2">
                        Higher values may take longer to process
                      </Paragraph>
                    )}
                  </div>
                </div>
              </>
            )}
            
            <div className="flex gap-4 mt-2">
              <PrimaryButton 
                onClick={handleSearch} 
                label="Search" 
              />
            </div>
          </div>
        </BorderBox>
      </div>
    </main>
  );
}