import Image from "next/image";
import { Subheading, SubheadingRed, DescriptionBlack, DescriptionWhite, DescriptionRed
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

export const ElementsCard = ({ name, imageSrc, className = "" }) => {
  return (
    <div
      className={`bg-secondary rounded-xl shadow-md p-4 w-[110px] h-[110px] flex flex-col items-center justify-center gap-3 ${className}`}
    >
      <Image src={imageSrc} alt={name} width={40} height={40} />
      <Subheading className="text-primary">{name}</Subheading>
    </div>
  );
};

export const ResultStatCard = ({ iconSrc, value, label, className = "" }) => {
  return (
    <div className="flex flex-col items-center gap-2">
        <Subheading className="text-secondary">{label}</Subheading>
      <div
        className={`bg-secondary text-center rounded-xl p-4 w-[130px] h-[140px] shadow-md border border-secondary flex flex-col items-center justify-between ${className}`}
      >
        <div className="h-[60px] flex items-center justify-center">
          <Image src={iconSrc} width={60} height={60} alt="Stat Icon" />
        </div>

        <div className="w-[130px] h-[2px] bg-primary" />

        <DescriptionRed>{value}</DescriptionRed>
      </div>
    </div>
  );
};

export const SearchTreeBox = ({ treeSrc }) => {
  return (
    <div className="bg-secondary rounded-xl p-6 w-full flex justify-center shadow-md">
      <div className="bg-background p-4 rounded">
        <Image src={treeSrc} alt="Tree Result" width={400} height={250} />
      </div>
    </div>
  );
};
