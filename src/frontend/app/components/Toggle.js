"use client";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import clsx from "clsx";

const RecipeToggle = () => {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const isShortest = pathname.includes("shortest");
  const algo = searchParams.get("algo") || "DFS"; // default fallback

  return (
    <div className="flex bg-secondary rounded-lg border border-secondary p-1 gap-5 w-fit">
      <button
        onClick={() => router.push(`/multiplerecipes?algo=${algo}`)}
        className={clsx(
          "px-8 py-2 font-bold text-sm rounded-lg font-poppins transition-all duration-150 cursor-pointer",
          !isShortest
            ? "bg-primary text-secondary border border-primary"
            : "text-primary" 
        )}
      >
        Multiple Recipes
      </button>

      <button
        onClick={() => router.push(`/shortestrecipe?algo=${algo}`)}
        className={clsx(
          "px-8 py-2 font-bold text-sm rounded-lg font-poppins transition-all duration-150 cursor-pointer",
          isShortest
            ? "bg-primary text-secondary border border-primary"
            : "text-primary"
        )}
      >
        Shortest Recipe
      </button>
    </div>
  );
};

export default RecipeToggle;