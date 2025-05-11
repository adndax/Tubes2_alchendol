"use client";
import { Suspense } from "react"; 
import { useSearchParams } from "next/navigation";
import { useRouter } from "next/navigation";
import { Heading, Paragraph } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";
import { PrimaryButton } from "@/components/Button";
import { ResultStatCard } from "@/components/Card";
import TreeDiagram from "@/components/TreeDiagram";
import { useState } from "react";

function ResultPageContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const target = searchParams.get("target");
  const algo = searchParams.get("algo") || "DFS";
  const mode = searchParams.get("mode") || "single";
  const quantity = parseInt(searchParams.get("quantity") || "1", 10);
  const [stats, setStats] = useState({ nodeCount: 0, timeMs: 0 });

  if (!target) {
    return <p className="text-red-500">❌ Target not specified.</p>;
  }

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-15 gap-15 w-full pb-20">
        <div className="flex flex-col gap-4 items-center">
          <Heading>Eureka! Here&apos;s Your Alchemy Route</Heading>
          <Paragraph>
            You searched, I conjured, and here it is — your magical recipe revealed!
          </Paragraph>
        </div>

        <BorderBox className="w-full">
          <div className="flex flex-col items-center p-10 gap-6">
            <TreeDiagram 
              target={target} 
              algo={algo}
              mode={mode === "multiple" ? "multiple" : "single"}
              maxRecipes={quantity}
              onStatsUpdate={({ nodeCount, timeMs }) => {
                setStats({ nodeCount, timeMs });
              }}
            />

            <div className="flex gap-10 mt-6">
              <ResultStatCard iconSrc="/img/time.png" value={`${stats.timeMs}ms`} label="Search Time" />
              <ResultStatCard iconSrc="/img/tree.png" value={`${stats.nodeCount} nodes`} label="Node Visited" />
            </div>
          </div>
        </BorderBox>

        <PrimaryButton label="Back To Home" onClick={() => router.push("/")} />
      </div>
    </main>
  );
}

export default function ResultPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <ResultPageContent />
    </Suspense>
  );
}