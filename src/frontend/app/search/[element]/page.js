"use client";
import { useParams, useSearchParams } from "next/navigation";
import { elements } from "@data";
import { Heading, Paragraph, Subheading } from "@/components/Typography";
import { ElementsCard } from "@/components/Card";
import { BorderBox } from "@/components/BorderBox";
import { PrimaryButton } from "@/components/Button";
import QuantityInput from "@/components/QuantityInput";
import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";

export default function ElementDetailPage() {
  const params = useParams();
  const element = decodeURIComponent(params.element);
  const searchParams = useSearchParams();
  const mode = searchParams.get("mode") || "shortest";
  const algo = searchParams.get("algo") || "dfs";
  const errorMsg = searchParams.get("error");
  
  const [data, setData] = useState(null);
  const [quantity, setQuantity] = useState(1);
  const [error, setError] = useState(errorMsg || null);
  
  const router = useRouter();
  
  // Find element data
  useEffect(() => {
    const foundElement = elements.find((el) => el.name === element);
    setData(foundElement);
    
    if (!foundElement) {
      setError("Element not found");
    }
  }, [element]);

  const handleSearch = () => {
    // Redirect to searching page with all parameters
    const isMultiple = mode === "multiple";
    const searchingUrl = `/searching?target=${encodeURIComponent(element)}&algo=${algo}&multiple=${isMultiple}${isMultiple ? `&maxRecipes=${quantity}` : ''}`;
    
    console.log("Navigating to searching page:", searchingUrl);
    router.push(searchingUrl);
  };

  if (!data) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        {error ? (
          <p className="text-center text-red-500">{error}</p>
        ) : (
          <p className="text-center">Loading element data...</p>
        )}
      </div>
    );
  }

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-28 gap-22 w-full pb-20">
        <div className="flex flex-col gap-2 items-center">
          <Heading>The Final Ingredient... Is You!</Heading>
          <Paragraph>
            The elements are ready… the scrolls have aligned… now all we need is a tiny click!
          </Paragraph>
          
          {error && (
            <Paragraph className="text-red-500 mt-4">
              {error}
            </Paragraph>
          )}
        </div>

        <BorderBox className="w-full">
          <div className="flex flex-col items-center p-10 gap-6">
            <ElementsCard name={data.name} imageSrc={data.imageSrc} />
            <Paragraph>{data.description}</Paragraph>
            {mode === "multiple" && (
              <>
                <div className="flex flex-col items-center gap-2">
                  <Subheading>Psst... how many recipes do you want?</Subheading>
                  <QuantityInput value={quantity} onChange={(val) => setQuantity(val)} />
                  <Paragraph className="text-sm text-gray-400">
                    (If fewer recipes exist than requested, all available recipes will be shown)
                  </Paragraph>
                </div>
              </>
            )}
          </div>
        </BorderBox>

        <PrimaryButton 
          onClick={handleSearch} 
          label="Search"
        />
      </div>
    </main>
  );
}