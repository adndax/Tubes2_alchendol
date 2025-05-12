"use client";
import { Suspense } from "react";  // Make sure this import is present
import { useState, useMemo } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Heading, Paragraph } from "@/components/Typography";
import RecipeToggle from "@/components/Toggle";
import { PrimaryButton } from "@/components/Button";
import SearchBar from "@/components/SearchBar";
import { ElementBox } from "@/components/BorderBox";
import { elements } from "@data";

function MultipleRecipesContent() {
  const searchParams = useSearchParams();
  const algo = searchParams.get("algo") || "null";
  
  const [searchQuery, setSearchQuery] = useState("");
  const [page, setPage] = useState(1);
  const [mode, setMode] = useState("multiple");

  const itemsPerRow = 8;
  const rowsPerPage = 5;
  const itemsPerPage = itemsPerRow * rowsPerPage;

  const filteredElements = useMemo(() => {
    return elements.filter((el) =>
      el.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [searchQuery]);

  const totalPages = Math.ceil(filteredElements.length / itemsPerPage);

  const currentElements = useMemo(() => {
    const start = (page - 1) * itemsPerPage;
    return filteredElements.slice(start, start + itemsPerPage);
  }, [page, filteredElements, itemsPerPage]);

  const handleSearch = (query) => {
    setSearchQuery(query);
    setPage(1); 
  };

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-20 gap-15 w-full pb-20">
        <div className="flex flex-col gap-2 items-center">
          <Heading>
            {algo !== "null"
              ? `You Picked the ${algo.toUpperCase()} Spell!`
              : "Pick Your Quest Mode"}
          </Heading>
          <Paragraph>
            Now… do you want a fast recipe or a magical recipe hunt?
          </Paragraph>
        </div>

        <RecipeToggle value={mode} onChange={setMode} />
        <SearchBar onSearch={handleSearch} />

        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-4 max-w-screen-xl">
          {currentElements.map((el, index) => (
            <ElementBox
              key={index}
              name={el.name}
              imageSrc={el.imageSrc}
              mode={mode}
              algo={algo}
            />
          ))}
        </div>

        <div className="flex items-center gap-5 pt-10">
          <PrimaryButton
            onClick={() => setPage((prev) => Math.max(prev - 1, 1))}
            disabled={page === 1}
            label="<"
          />
          <span className="font-poppins font-bold text-secondary">
            Page {page} of {totalPages}
          </span>
          <PrimaryButton
            onClick={() => setPage((prev) => Math.min(prev + 1, totalPages))}
            disabled={page === totalPages}
            label=">"
          />
        </div>
      </div>
    </main>
  );
}

export default function MultipleRecipes() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <MultipleRecipesContent />
    </Suspense>
  );
}