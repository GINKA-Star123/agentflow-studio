import type { Metadata } from "next";

import {Providers} from "@/app/providers";

import "./globals.css";

export const metadata:Metadata = {
  title:"AgentFlow Studio",
  description:"AI Workflow 平台",
};

type RootLayoutProps = {
  children:React.ReactNode;
};

export default function RootLayout({children}:RootLayoutProps){
  return(
    <html lang="zh-CN">
      <body>
        <Providers>{children}</Providers>
        </body>
    </html>
  )
}