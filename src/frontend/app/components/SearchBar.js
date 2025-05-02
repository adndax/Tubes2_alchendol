import { useState } from "react";
import { FiSearch } from "react-icons/fi";

const SearchBar = ({ placeholder = "Search", onSearch }) => {
  const [query, setQuery] = useState("");

  const handleSubmit = (e) => {
    e.preventDefault();
    onSearch?.(query); 
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="flex items-center bg-white rounded-full px-4 py-2.5 w-full max-w-xl shadow-md"
    >
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder={placeholder}
        className="flex-grow bg-transparent outline-none font-poppins text-black text-sm placeholder:text-gray-500 font-medium"
      />
      <button type="submit" className="text-black cursor-pointer hover:text-gray-500">
        <FiSearch size={20} />
      </button>
    </form>
  );
};

export default SearchBar;