import Image from "next/image";
import { SubheadingRed, Subheading } from "./Typography";
import { DescriptionBlack, DescriptionWhite } from "./Typography";

const SpellCard = ({ imageSrc, title, subtitle, isSelected, onClick }) => {
  return (
    <div
      className={`
        ${isSelected ? "bg-primary text-secondary" : "bg-secondary text-foreground"} 
        rounded-xl shadow-md p-6 text-center w-full max-w-3xs cursor-pointer 
        transition-all duration-200
      `}
      onClick={onClick}
    >
      <div className="flex flex-col justify-center items-center gap-2">
        <Image src={imageSrc} alt={title} width={180} height={180} />
        {isSelected ? <Subheading>{title}</Subheading> : <SubheadingRed>{title}</SubheadingRed>}
        {isSelected ? <DescriptionWhite>{subtitle}</DescriptionWhite> : <DescriptionBlack>{subtitle}</DescriptionBlack>}
      </div>
    </div>
  );
};

export default SpellCard;