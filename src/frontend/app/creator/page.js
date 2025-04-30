"use client";
import Image from "next/image";
import { FloatingNav } from "@/components/Navbar";
import { Heading, Paragraph, Subheading } from "@/components/Typography";
import { BorderBox } from "@/components/BorderBox";

export default function Creator() {
  return (
    <main className="min-h-screen h-225 bg-background flex flex-col items-center p-8 text-foreground font-body">
        <FloatingNav/>

        <div className="flex flex-col items-center pt-40 gap-5 w-full pb-50">
          <Heading>
            Who’s Behind the Magical Cendol?
          </Heading>
          <div className="flex flex-row items-center gap-11">
            <div className="flex flex-col items-center">
              <Subheading>Muhammad Alfansya</Subheading>
              <p>13523005</p>
            </div>
            <Image
            src="/img/alchemist.png"
            alt="Alchendol Logo"
            width={280}
            height={280}
            />
            <div className="flex flex-col items-center">
              <Subheading>M. Hazim R. Prajoda</Subheading>
              <p>13523009</p>
            </div>
          </div>
          <div className="flex flex-col items-center">
              <Subheading>Adinda Putri</Subheading>
              <p>13523071</p>
          </div>
        </div>

        <div className="flex flex-col items-center gap-15 w-full pb-50">
          <Heading>
            Just a Sprinkle of Chaos & Code
          </Heading>
        <BorderBox>
          <div className="w-100 text-center mx-auto py-15 md:w-170">
          <Paragraph>Tralalaaaa! We’re a bunch of enthusiastic Informatics Engineer students from Bandung Institute of Technology who decided to mix Alchemy and Cendol into one delicious project!
          <br/><br/>
          Fueled by caffeine, curiosity, and a touch of chaos, Alchendol was brewed to help you uncover all the quirky item combos in Little Alchemy 2!
          <br/><br/>
          We hope this app brings you joy, surprises, and a sprinkle of giggles
          <br/><br/>
          — All of us, probably</Paragraph>
          </div>
        </BorderBox>
        </div>

    </main>
  );
}