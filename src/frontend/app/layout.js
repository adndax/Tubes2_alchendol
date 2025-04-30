import { Baloo_2, Poppins } from "next/font/google";
import "./globals.css";

const baloo = Baloo_2({
  subsets: ["latin"],
  variable: "--font-baloo",
  display: "swap",
});

const poppins = Poppins({
  subsets: ["latin"],
  variable: "--font-poppins",
  weight: ["400", "500", "600"],
  display: "swap",
});

export const metadata = {
  title: "Your App Title",
  description: "Your app description",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <head>
        <link
          href="https://fonts.googleapis.com/css2?family=Baloo+2&family=Poppins&display=swap"
          rel="stylesheet"
        />
      </head>
      <body
        className={`${baloo.variable} ${poppins.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}
