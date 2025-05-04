"use client"; // ⬅️ ini penting
import { Heading, Paragraph } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";
import { useRouter } from "next/navigation";
import { PrimaryButton } from "@/components/Button";
import { SearchTreeBox, ResultStatCard } from "@/components/Card";


export default function ResultPage() {
const router = useRouter();
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
          <SearchTreeBox treeSrc="/img/result.png" />

          <div className="flex gap-10 mt-6">
            <ResultStatCard iconSrc="/img/time.png" value="120ms" label="Search Time"/>
            <ResultStatCard iconSrc="/img/tree.png" value="100 nodes" label="Node Visited"/>
          </div>
        </div>
      </BorderBox>

      <PrimaryButton label="Back To Home" onClick={() => router.push("/")}/>

      </div>
    </main>
  );
}