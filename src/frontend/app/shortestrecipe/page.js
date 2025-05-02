"use client";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Heading, Paragraph } from "@/components/Typography";
import RecipeToggle from "@/components/Toggle";
import { PrimaryButton } from "@/components/Button";
import SearchBar from "@/components/SearchBar";

export default function Greets() {
  const router = useRouter();
  const handleSearch = (query) => {
    console.log("Search for:", query);
  };

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-20 gap-15 w-full pb-20">
        <div className="flex flex-col gap-2 items-center">
            <Heading>Pick Your Quest Mode</Heading>
            <Paragraph>
                Now… do you want a fast recipe or a magical recipe hunt?
            </Paragraph>
        </div>
        <RecipeToggle />
        <SearchBar onSearch={handleSearch} />
        <PrimaryButton onClick={() => router.push("/howtoplay")} label="Meow" />
      </div>
    </main>
  );
}