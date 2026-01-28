import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { getAuthenticatedUser } from "@/lib/auth";
import { AuthProvider } from "./providers/auth-provider";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "StoreOS Admin",
  description: "Admin dashboard for managing the application",
};

export default  function AdminLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  

  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
     
          {children}
        
      </body>
    </html>
  );
}
