"use client";
import { cn } from "@utils";

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