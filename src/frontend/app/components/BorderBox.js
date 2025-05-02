"use client";
import Image from "next/image";
import { cn } from "@utils";
import { Subheading } from "./Typography";

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

export const ElementBox = ({ name, imageSrc }) => {
  return (
    <div className="flex flex-col items-center justify-center border border-secondary rounded-md p-2 w-29 h-29 bg-background gap-3">
      <Image src={imageSrc} alt={name} width={40} height={40} />
      <Subheading>{name}</Subheading>
    </div>
  );
};