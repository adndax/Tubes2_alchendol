"use client";
import React, { useState } from "react";
import {
  motion,
  AnimatePresence,
  useScroll,
  useMotionValueEvent,
} from "framer-motion";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { navItems } from "@data";  
import { cn } from "@utils";


export const FloatingNav = ({ className }) => {
  const { scrollYProgress } = useScroll();
  const [visible, setVisible] = useState(true);
  const pathname = usePathname(); 

  useMotionValueEvent(scrollYProgress, "change", (current) => {
    if (typeof current === "number") {
      let direction = current - scrollYProgress.getPrevious();

      if (scrollYProgress.get() < 0.05) {
        setVisible(true);
      } else {
        if (direction < 0) {
          setVisible(true);
        } else {
          setVisible(false);
        }
      }
    }
  });
  
  // to do: hamburger menu buat page responsive
  return (
    <AnimatePresence mode="wait">
      <motion.div
        initial={{ opacity: 1, y: -100 }}
        animate={{ y: visible ? 0 : -100, opacity: visible ? 1 : 0 }}
        transition={{ duration: 0.2 }}
        className={cn(
            "flex fixed top-8 left-8 right-8 z-[5000] bg-secondary rounded-xl shadow-lg max-w-4xl mx-auto py-1",
            className
          )}
        >
          {navItems.map((item) => {
            const isActive = pathname === item.link;
            return (
              <div 
                key={item.link}
                className="flex-1 flex justify-center"
              >
                <Link
                  href={item.link}
                  className={cn(
                    "inline-block text-center text-sm font-semibold transition-colors duration-200 font-poppins mx-1",
                    isActive ? "bg-primary text-secondary rounded-lg px-6 py-2" : "text-primary hover:text-primary-hover px-6 py-2"
                  )}
                >
                  {item.name}
                </Link>
              </div>
            );
          })}
        </motion.div>
    </AnimatePresence>
  );
};