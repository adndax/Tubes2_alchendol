"use client";
import { useState, useEffect } from "react";
import { SubheadingRed } from "./Typography";

export default function QuantityInput({ value = 1, onChange }) {
  const [quantity, setQuantity] = useState(value);


  const update = (val) => {
    const num = Math.max(1, val); // Changed from 0 to 1 for minimum value
    setQuantity(num);
    onChange?.(num);
  };

  // Ensure the initial value is at least 1
  useEffect(() => {
    if (value < 1) {
      update(1);
    }
  }, [value, update]);

  const handleInput = (e) => {
    const val = e.target.value;
  
    if (val === "") {
      setQuantity("");
      // When empty, use default value of 1
      onChange?.(1);
      return;
    }
  
    const num = parseInt(val, 10);
    if (!isNaN(num)) update(num);
  };

  return (
    <div className="flex items-center bg-secondary rounded-lg border-2 border-primary px-4 py-1 gap-4 w-fit">
      <button 
        onClick={() => update(quantity - 1)} 
        className="cursor-pointer relative transition-shadow duration-200 hover:shadow-md"
      >
        <span className="absolute -inset-3" aria-hidden="true" />
        <SubheadingRed>-</SubheadingRed>
      </button>
      <input
        type="number"
        min="1" // Changed from 0 to 1
        value={quantity}
        onChange={handleInput}
        className="bg-transparent text-center w-7 outline-none font-bold text-primary 
        appearance-none 
        [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none
        [&::-webkit-inner-spin-button]:m-0 [&::-webkit-outer-spin-button]:m-0
        [&::-webkit-inner-spin-button]:opacity-0 [&::-webkit-outer-spin-button]:opacity-0
        [-moz-appearance:textfield]"
      />
      <button 
        onClick={() => update(quantity + 1)} 
        className="cursor-pointer relative transition-shadow duration-200 hover:shadow-md"
      >
        <span className="absolute -inset-3" aria-hidden="true" />
        <SubheadingRed>+</SubheadingRed>
      </button>
    </div>
  );
}