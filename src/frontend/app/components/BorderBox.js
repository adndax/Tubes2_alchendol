"use client";
import Image from "next/image";
import { cn } from "@utils";
import { Subheading } from "./Typography";
import { useRouter } from "next/navigation";

export const BorderBox = ({ children, className, height }) => (
  <div className="w-full flex justify-center">
    <div
      className={cn(
        "relative border border-secondary rounded-xl bg-background text-foreground overflow-hidden",
        "w-full max-w-4xl",
        height,
        className
      )}
    >
      {children}
    </div>
  </div>
);

export const ElementBox = ({ name, imageSrc, mode = "shortest", algo = "DFS" }) => {
  const router = useRouter();

  // Fix for handling null algorithm
  let safeAlgo = algo;
  if (safeAlgo === "null" || safeAlgo === null || safeAlgo === undefined) {
    safeAlgo = "DFS";
  }

  const handleClick = () => {
    // Always use a valid algorithm and properly encode parameters
    const url = `/search/${encodeURIComponent(name)}?mode=${encodeURIComponent(mode)}&algo=${encodeURIComponent(safeAlgo)}`;
    
    // Debug - log the URL we're navigating to
    console.log("ElementBox navigating to:", url);
    
    router.push(url);
  };

  return (
    <div
      onClick={handleClick}
      className="cursor-pointer flex flex-col items-center justify-center border border-secondary rounded-md p-2 w-29 h-29 bg-background gap-3 hover:shadow-md transition"
    >
      <Image src={imageSrc} alt={name} width={40} height={40} />
      <Subheading>{name}</Subheading>
    </div>
  );
};