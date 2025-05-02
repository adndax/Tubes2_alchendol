import { cn } from "@utils";

export const Heading = ({ children, className }) => (
  <h1 className={cn("text-3xl font-bold text-secondary font-baloo text-center", className)}>
    {children}
  </h1>
);

export const Subheading = ({ children, className }) => (
  <h2 className={cn("font-bold text-secondary font-baloo text-center", className)}>
    {children}
  </h2>
);

export const SubheadingRed = ({ children, className }) => (
  <h2 className={cn("font-bold text-primary font-baloo text-center", className)}>
    {children}
  </h2>
);

export const Step = ({ children, className }) => (
  <h2 className={cn("text-2xl font-bold text-white font-baloo text-left", className)}>
    {children}
  </h2>
);

export const Paragraph = ({ children, className }) => (
  <p className={cn("text-white text-sm font-poppins font-medium text-center", className)}>
    {children}
  </p>
);

export const DescriptionBlack= ({ children, className }) => (
  <p className={cn("text-black text-xs font-poppins font-medium text-center", className)}>
    {children}
  </p>
);

export const DescriptionWhite= ({ children, className }) => (
  <p className={cn("text-white text-xs font-poppins font-medium text-center", className)}>
    {children}
  </p>
);

export const Information = ({ children, className }) => (
  <p className={cn("text-white text-sm font-poppins font-medium text-center", className)}>
    {children}
  </p>
);