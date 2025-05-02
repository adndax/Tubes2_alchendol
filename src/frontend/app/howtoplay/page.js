"use client";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Heading, Paragraph, Step, Subheading } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";
import { PrimaryButton } from "@/components/Button";

export default function HowToPlay() {
    const router = useRouter();
    return (
        <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
            <div className="flex flex-col items-center pt-10 w-full pb-20">
                <div className="overflow-visible">
                    <div className="relative">
                        <div className="flex flex-col gap-2">
                            <Heading>
                                How To Play - Guided by Meowculus!
                            </Heading>
                            <Paragraph>
                                Worry not, curious crafter! I’ll guide you through this alchemical adventure step-by-step. 
                                <br/>Let’s cook up some magic!
                            </Paragraph>
                        </div>
                        <Image
                        src="/img/sparkle.png"
                        alt="Meowculus"
                        width={100}
                        height={100}
                        className="absolute -top-7 -left-25 z-10"
                        />
                        <Image
                        src="/img/meowculus_confused.png"
                        alt="Meowculus"
                        width={120}
                        height={120}
                        className="absolute -bottom-17 -right-30 z-10"
                        />
                    </div>
                </div>
            </div>

            <div className="flex flex-col w-full pb-30">
                <div className="flex flex-row items-center justify-center">
                    <Image
                        src="/img/compass.png"
                        alt="Meowculus"
                        width={60}
                        height={60}
                        />
                    <Step>
                        Step 1: Choose Your Magic Path
                    </Step>
                </div>
                <BorderBox className="w-full">
                    <div className="flex flex-col gap-5 p-7 px-5">
                        <Paragraph>
                            Every wizard needs a method. Pick the search spells!
                        </Paragraph>
                        <div className="flex flex-row justify-center gap-10">
                            <div className="flex flex-col items-center py-2 gap-2">
                                <Image
                                    src="/img/bfs.png"
                                    alt="Meowculus"
                                    width={120}
                                    height={120}
                                />
                                <Subheading>
                                    BFS (Breadth-First Spell)
                                </Subheading>
                            </div>
                            <div className="flex flex-col items-center py-3 gap-2">
                                <Image
                                    src="/img/bidirectional.png"
                                    alt="Meowculus"
                                    width={120}
                                    height={120}
                                />
                                <Subheading>
                                    Bidirectional Beam
                                </Subheading>
                            </div>
                            <div className="flex flex-col items-center gap-2">
                                <Image
                                    src="/img/dfs1.png"
                                    alt="Meowculus"
                                    width={120}
                                    height={120}
                                />
                                <Subheading>
                                    DFS (Depth-Focused Sorcery)
                                </Subheading>
                            </div>
                        </div>
                    </div>
                    
                </BorderBox>
            </div>

            <div className="flex flex-col w-full pb-30">
                <div className="flex flex-row items-center justify-center">
                    <Step>
                        Step 2: Pick Your Crafting Style
                    </Step>
                    <Image
                        src="/img/atom.png"
                        alt="Meowculus"
                        width={60}
                        height={60}
                        />
                </div>
                <BorderBox className="w-full">
                    <div className="flex flex-col gap-5 py-7 px-5">
                        <Paragraph>
                            Now, do you want to find the quickest way to your target… 
                            <br/>
                            or explore many magical paths?
                        </Paragraph>
                        <div className="flex flex-row justify-center gap-20">
                                <Image
                                    src="/img/shortest_recipe.png"
                                    alt="Meowculus"
                                    width={180}
                                    height={180}
                                />
                                <Image
                                    src="/img/multiple_recipe.png"
                                    alt="Meowculus"
                                    width={180}
                                    height={180}
                                />
                        </div>
                    </div>
                    
                </BorderBox>
            </div>

            <div className="flex flex-row gap-10 max-w-225">
            <div className="flex flex-col w-full pb-30 gap-3">
                <div className="flex flex-row items-center justify-center">
                    <Image
                        src="/img/search.png"
                        alt="Meowculus"
                        width={60}
                        height={60}
                        />
                    <Step>
                        Step 3A: Shortest Recipe (Quick & Neat)
                    </Step>
                </div>
                <BorderBox className="w-full">
                    <div className="flex flex-col py-7 px-5">
                    <Paragraph>
                        Now tell me… 
                        <br></br>
                        what treasure are we crafting today?
                        <br></br>
                        1. Type or select the element you want to find 
                        <br/><br/>
                        2. Hit Search and POOF! — your recipe shall appear!
                    </Paragraph>
                    </div>
                    
                </BorderBox>
            </div>

            <div className="flex flex-col w-full pb-30 gap-4">
                <div className="flex flex-row items-center justify-center">
                    <Image
                        src="/img/cardboard1.png"
                        alt="Meowculus"
                        width={60}
                        height={60}
                        />
                    <Step>
                        Step 3B: Multiple Recipes (The More, the Merrier!)
                    </Step>
                </div>
                <BorderBox className="w-full">
                    <div className="flex flex-col py-7 px-5">
                        <Paragraph>
                            Ahh, a true explorer of possibility!
                            <br/><br/>   
                            1. Choose your desired element 
                            <br/><br/>
                            2. Enter how many recipes you wish to uncover <br/> (ex: 3? 7? 99?) 
                            <br/><br/>
                            3. Tap Search, and let the quest begin!
                        </Paragraph>
                    </div>
                    
                </BorderBox>
            </div>
            </div>


            <div className="flex flex-col w-full pb-30 gap-3">
                <div className="flex flex-row items-center justify-center">
                    <Step>
                        Step 4: Behold the Alchemical Tree
                    </Step>
                </div>
                <BorderBox className="w-full">
                    <div className="flex flex-col gap-5 py-7 px-5">
                        <Paragraph>
                            Ta-da! Your magical recipe map will appear!
                        </Paragraph>
                        <div className="flex flex-row justify-center gap-10 items-center">
                            <div className="flex flex-col items-center gap-4 pt-1">
                                <Image
                                    src="/img/tree.png"
                                    alt="Meowculus"
                                    width={110}
                                    height={110}
                                />
                                <Paragraph>
                                    A recipe tree showing how basic elements become your goal
                                </Paragraph>
                            </div>
                            <div className="flex flex-col items-center pt-1">
                                <Image
                                    src="/img/time.png"
                                    alt="Meowculus"
                                    width={90}
                                    height={90}
                                />
                                <Paragraph>
                                    The search time it took — blink and you might miss it!
                                </Paragraph>
                            </div>
                            <div className="flex flex-col items-center gap-3">
                                <Image
                                    src="/img/cardboard2.png"
                                    alt="Meowculus"
                                    width={120}
                                    height={120}
                                />
                                <Paragraph>
                                    The number of nodes visited — how deep did your spell dive?
                                </Paragraph>
                            </div>
                        </div>
                    </div>
                </BorderBox>
            </div>
                <div className="pb-10">
                    <PrimaryButton onClick={() => router.push("/magicpath")} label="Meow"/>
                </div>
        </main>
    );
}