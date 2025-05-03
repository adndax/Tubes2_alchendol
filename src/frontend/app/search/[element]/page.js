"use client"; // ⬅️ ini penting
import { useParams } from "next/navigation";
import { elements } from "@data";
import { Heading, Paragraph, Subheading } from "@/components/Typography";
import { ElementsCard } from "@/components/Card";
import { BorderBox } from "@/components/BorderBox";
import { PrimaryButton } from "@/components/Button";
import QuantityInput from "@/components/QuantityInput";
import { useSearchParams } from "next/navigation";

export default function ElementDetailPage() {
  const params = useParams();
  const element = decodeURIComponent(params.element);
  const searchParams = useSearchParams();
  const mode = searchParams.get("mode") || "shortest"; 
  const data = elements.find((el) => el.name === element);

  if (!data) {
    return <p className="text-center text-red-500">Element not found</p>;
  }

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
    <div className="flex flex-col items-center pt-20 gap-22 w-full pb-20">
      <div className="flex flex-col gap-2 items-center">
        <Heading>The Final Ingredient... Is You!</Heading>
        <Paragraph>
        The elements are ready… the scrolls have aligned… now all we need is a tiny click!
        </Paragraph>
      </div>

      <BorderBox className="w-full">
        <div className="flex flex-col items-center p-10 gap-6">
        <ElementsCard name={data.name} imageSrc={data.imageSrc} />
        <Paragraph>{data.description}</Paragraph>
        <div className="flex flex-col items-center gap-2">
            <Subheading>
                Psst... how many do you want?
            </Subheading>
            {mode === "multiple" && (
              <QuantityInput value={0} onChange={(val) => console.log("Jumlah:", val)} />
              )}
        </div>
        </div>
      </BorderBox>

      <PrimaryButton label="Search"/>

      </div>
    </main>
  );
}