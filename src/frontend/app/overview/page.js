"use client";
import Image from "next/image";
import { FloatingNav } from "@/components/Navbar";
import { Heading, Paragraph } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";

export default function Creator() {
  return (
    <main className="min-h-screen h-225 bg-background flex flex-col items-center p-8 text-foreground font-body">
        <FloatingNav/>
        
        <div className="flex flex-col items-center pt-50 gap-15 w-full pb-40">
          <Heading>
            Where Alchemy Meets... Cendol?
          </Heading>
          <BorderBox>
            <div className="flex flex-col md:flex-row items-center justify-center gap-6 px-8 py-10">
                <Paragraph className="max-w-xl text-left">
                    <span className="font-bold text-secondary italic">Welcome to Alchendol!</span> — a magical helper built just for you, the curious crafter of Little Alchemy 2!
                    <br /><br />
                    We created this site so you don’t have to guess your way to Unicorns or Bricks anymore.
                    <br /><br />
                    Just type in your desired item, choose an algorithm (BFS, DFS, Bidirectional), and voilà! Watch as recipe trees bloom before your eyes!
                </Paragraph>
                <Image
                    src="/img/alchendol_logo.png"
                    alt="Alchendol Logo"
                    width={200}
                    height={200}
                    className="flex-shrink-0"
                    />
            </div>
        </BorderBox>
        </div>

        <div className="flex flex-col items-center w-full pb-50">
            <Image
                src="/img/meowculus_hi.png"
                alt="Alchendol Logo"
                width={200}
                height={200}
                className="flex-shrink-0"
                />
            <div className="flex flex-col gap-10 w-full">
                <Heading>
                    Meet Meowculus, Your Potion Pal!
                </Heading>
                <BorderBox>
                    <div className="flex flex-col items-center gap-6 px-8 py-10">
                        <Paragraph className="max-w-xl">
                        Say hello to <span className="font-bold text-secondary italic">Meowculus</span> , our mischievous magical cat alchemist!
                        <br/><br/>
                        He’s here to guide your journey, stir the pot, ask the silly questions, and occasionally… explode things by accident 
                        <br/><br/>
                        From curious thinking to joyful discovery, Meowculus is the spirit of Alchendol — wise, playful, and slightly caffeinated.
                        </Paragraph>
                    </div>
                </BorderBox>
            </div>
        </div>

        <div className="flex flex-row items-center justify-center gap-10 w-full pb-50">
            <div className="flex flex-col gap-5 items-start">
                <Heading>
                    Watch the Magic Happen!
                </Heading>
                <Paragraph className="text-start">
                    Still wondering how Alchendol works? Don’t worry, we’ve brewed a <span className="font-bold text-secondary italic">demo video</span> just for you!
                    <br/><br/>
                    Watch as items combine, trees grow, and chaos unfolds — all with just one click on Search
                </Paragraph>
            </div>
            <Image
                src="/img/meowculus_watching.png"
                alt="Alchendol Logo"
                width={200}
                height={200}
                className="flex-shrink-0"
                />
        </div>

    </main>
  );
}