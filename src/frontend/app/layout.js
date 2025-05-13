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

// Metadata yang akan dipakai untuk SEMUA halaman
export const metadata = {
  title: "Alchendol", // Judul yang sama untuk semua halaman
  description: "Website for researching recipes in Little Alchemy 2",
  icons: {
    icon: [
      { url: "/img/alchendol_logo.png" }, 
      { url: "/favicon.ico" }
    ],
    apple: { url: "/img/alchendol_logo.png" },
  },
  manifest: "/manifest.json",
  applicationName: "Alchendol",
  appleWebApp: { capable: true, title: "Alchendol", statusBarStyle: "default" },
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body className={`${baloo.variable} ${poppins.variable} antialiased`}>
        {children}
      </body>
    </html>
  );
}