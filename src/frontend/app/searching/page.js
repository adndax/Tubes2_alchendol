"use client"; // ⬅️ ini penting
import { SecondaryButton } from "@/components/Button";
import { Heading, Paragraph, Subheading } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";
import Image from "next/image";
import { useRouter } from "next/navigation";


export default function ElementDetailPage() {
    const router = useRouter();
    
    return (
        <main className="min-h-screen bg-background flex flex-col items-center p-8 text-foreground font-body">
        <div className="flex flex-col items-center pt-28 gap-22 w-full pb-20">
        <div className="flex flex-col gap-8 items-center">
            <Heading>Brewing the Recipe…</Heading>
            <Paragraph>
                Meowculus is sniffing through the archives…
            </Paragraph>
        </div>

        <BorderBox className="w-full">
            <div className="flex flex-col items-center p-10 gap-6">
                <Image
                    src="/img/meowculus_stir.png"
                    alt="Meowculus"
                    width={110}
                    height={110}
                />
                <Paragraph>
                    Searching...
                </Paragraph>
            </div>
        </BorderBox>


        <SecondaryButton onClick={() => router.push("/result")} label="Cancel"/>

        </div>
        </main>
    );
    }