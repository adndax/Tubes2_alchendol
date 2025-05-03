import Image from "next/image";
import { Subheading, SubheadingRed, DescriptionBlack, DescriptionWhite,
} from "./Typography";

export const SpellCard = ({ imageSrc, title, subtitle, isSelected, onClick }) => {
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
        {isSelected ? (
          <DescriptionWhite>{subtitle}</DescriptionWhite>
        ) : (
          <DescriptionBlack>{subtitle}</DescriptionBlack>
        )}
      </div>
    </div>
  );
};

export const ElementsCard = ({ name, imageSrc, onClick }) => {
  return (
    <div
      onClick={onClick}
      className="bg-secondary rounded-xl shadow-md p-4 w-[110px] h-[110px] flex flex-col items-center justify-center gap-3"
    >
      <Image src={imageSrc} alt={name} width={40} height={40} />
      <Subheading className="text-primary">{name}</Subheading>
    </div>
  );
};