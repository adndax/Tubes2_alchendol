"use client";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Heading, Paragraph } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";
import { PrimaryButton } from "@/components/Button";

export default function Greets() {
    const router = useRouter();
    return (
        <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
            <div className="flex flex-col items-center pt-50 gap-15 w-full pb-40">
                <Heading>
                    Greets!
                </Heading>
                <BorderBox className="overflow-visible">
                    <div className="relative">
                        <div className="flex flex-col items-center gap-6 pl-10 pr-30 py-10">
                        <Paragraph className="max-w-xl text-left">
                            Meowgical greetings, explorer! I’m 
                            <span className="font-bold text-secondary italic">Meowculus</span>, your potion partner in the world of Alchendol! Ready to brew some crazy combos? Let’s mix, match, and discover the unexpected!
                        </Paragraph>
                        </div>

                        <Image
                        src="/img/meowculus_happy.png"
                        alt="Meowculus"
                        width={180}
                        height={180}
                        className="absolute -bottom-20 -right-11 z-10"
                        />
                    </div>
                </BorderBox>
                <PrimaryButton onClick={() => router.push("/howtoplay")} label="Meow"/>
            </div>
        </main>
    );
}