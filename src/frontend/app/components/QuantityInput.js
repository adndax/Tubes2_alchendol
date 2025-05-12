"use client";
import { useState } from "react";
import { SubheadingRed } from "./Typography";
import { useEffect } from "react";

export default function QuantityInput({ value = 1, onChange }) {
  const [quantity, setQuantity] = useState(value);

  const update = (val) => {
    const num = Math.max(1, val); // Minimal value 1, bukan 0
    setQuantity(num);
    onChange?.(num);
  };

  const handleInput = (e) => {
    const val = e.target.value;
  
    if (val === "") {
      setQuantity("");
      onChange?.(1); // Set default value to 1 when empty
      return;
    }
  
    const num = parseInt(val, 10);
    if (!isNaN(num)) update(num);
  };

  // Pastikan nilai awal minimal 1
  useEffect(() => {
    if (value < 1) {
      update(1);
    }
  }, []);

  return (
    <div className="flex items-center bg-secondary rounded-lg border-2 border-primary px-4 py-1 gap-4 w-fit">
      <button onClick={() => update(quantity - 1)} className="cursor-pointer">
        <SubheadingRed>-</SubheadingRed>
      </button>
      <input
        type="number"
        min="1"
        value={quantity}
        onChange={handleInput}
        className="bg-transparent text-center w-7 outline-none font-bold text-primary appearance-none [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
        />
      <button onClick={() => update(quantity + 1)} className="cursor-pointer">
        <SubheadingRed>+</SubheadingRed>
      </button>
    </div>
  );
}