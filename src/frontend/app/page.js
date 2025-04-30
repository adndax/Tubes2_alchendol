"use client";
import Image from "next/image";
import { FloatingNav } from "@/components/Navbar";
import { PrimaryButton } from "./components/Button";
import { useRouter } from "next/navigation";

export default function Page() {
  const router = useRouter();
  return (
    <main className="min-h-screen h-225 bg-background flex flex-col items-center p-8 text-foreground font-body">
        <FloatingNav/>

        <div className="mt-50 mb-5">
            <Image
            src="/img/alchendol_logoname.png"
            alt="Alchendol Logo"
            width={300}
            height={300}
            />
        </div>
        <PrimaryButton onClick={() => router.push("/greets")} label="Start"/>
    </main>
  );
}