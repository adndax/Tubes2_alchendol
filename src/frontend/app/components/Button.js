export const PrimaryButton = ({ onClick, label }) => {
    return (
      <button
        onClick={onClick}
        className="bg-primary text-secondary border-2 border-secondary px-7 py-1 rounded-lg font-bold tracking-widest uppercase font-baloo cursor-pointer
          hover:bg-primary-hover active:scale-95 transition-all duration-150"
      >
        {label}
      </button>
    );
  };
  
  export const SecondaryButton = ({ onClick, label }) => {
    return (
      <button
        onClick={onClick}
        className="bg-secondary text-primary border-2 border-primary px-7 py-1 rounded-lg font-bold tracking-widest uppercase font-baloo cursor-pointer
          hover:bg-secondary-hover active:scale-95 transition-all duration-150"
      >
        {label}
      </button>
    );
  };