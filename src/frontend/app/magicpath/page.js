"use client";
import { useState } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Heading, Paragraph } from "@/components/Typography";
import { SpellCard } from "@/components/Card";
import { spellCards } from "@data";
import { PrimaryButton } from "@/components/Button";

export default function MagicPath() {
  const router = useRouter();
  const [selectedCardIndex, setSelectedCardIndex] = useState(null);

  return (
    <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
      <div className="flex flex-col items-center pt-20 gap-15 w-full pb-20">
        <div className="flex flex-row gap-5 items-center">
          <div className="flex flex-col gap-2">
            <Heading>Choose Your Magic Path</Heading>
            <Paragraph>
              Every wizard needs a method. Pick one of my three favorite search spells!
            </Paragraph>
          </div>
          <Image src="/img/meowculus_ask.png" alt="Meowculus" width={120} height={120} />
        </div>

        <div className="flex flex-col md:flex-row gap-15 justify-center">
          {spellCards.map((card, index) => (
            <SpellCard
              key={index}
              {...card}
              isSelected={selectedCardIndex === index}
              onClick={() => setSelectedCardIndex(index === selectedCardIndex ? null : index)}
            />
          ))}
        </div>

        <PrimaryButton onClick={() => router.push("/multiplerecipes")} label="Meow" />
      </div>
    </main>
  );
}